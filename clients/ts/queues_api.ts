export class QueuesAPIError extends Error {
    public readonly code: number;
    public readonly details: any;

    constructor(message: string, code: number, details: any) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}


export interface Queue {
    name: string
    target: string
    retry: number
    interval: JsonDuration
}

export type JsonDuration = string; // suffixes: ns, us, ms, s, m, h

export type Token = string;




// support stuff


interface rpcExecutor {
    call(id: number, payload: string): Promise<object>;
}

class wsExecutor {
    private socket?: WebSocket;
    private connecting = false;
    private readonly pendingConnection: Array<() => (void)> = [];
    private readonly correlation = new Map<number, [(data: object) => void, (err: object) => void]>();

    constructor(private readonly url: string) {
    }

    async call(id: number, payload: string): Promise<object> {
        const conn = await this.connectIfNeeded();
        if (this.correlation.has(id)) {
            throw new Error(`already exists pending request with id ${id}`);
        }
        let future = new Promise<object>((resolve, reject) => {
            this.correlation.set(id, [resolve, reject]);
        });
        conn.send(payload);
        return (await future);
    }

    private async connectIfNeeded(): Promise<WebSocket> {
        while (this.connecting) {
            await new Promise((resolve => {
                this.pendingConnection.push(resolve);
            }))
        }
        if (this.socket) {
            return this.socket;
        }
        this.connecting = true;
        let socket;
        try {
            socket = await this.connect();
        } finally {
            this.connecting = false;
        }
        socket.onerror = () => {
            this.onConnectionFailed();
        }
        socket.onclose = () => {
            this.onConnectionFailed();
        }
        socket.onmessage = ({data}) => {
            let res;
            try {
                res = JSON.parse(data);
            } catch (e) {
                console.error("failed parse request:", e);
            }
            const task = this.correlation.get(res.id);
            if (task) {
                this.correlation.delete(res.id);
                task[0](res);
            }
        }
        this.socket = socket;

        let cp = this.pendingConnection;
        this.pendingConnection.slice(0, 0);
        cp.forEach((f) => f());
        return this.socket;
    }

    private connect(): Promise<WebSocket> {
        return new Promise<WebSocket>(((resolve, reject) => {
            let socket = new WebSocket(this.url);
            let resolved = false;
            socket.onopen = () => {
                resolved = true;
                resolve(socket);
            }

            socket.onerror = (e) => {
                if (!resolved) {
                    reject(e);
                    resolved = true;
                }
            }

            socket.onclose = (e) => {
                if (!resolved) {
                    reject(e);
                    resolved = true;
                }
            }
        }));
    }

    private onConnectionFailed() {
        let sock = this.socket;
        this.socket = undefined;
        if (sock) {
            sock.close();
        }
        const cp = Array.from(this.correlation.values());
        this.correlation.clear();
        const err = new Error('connection closed');
        cp.forEach((([_, reject]) => {
            reject(err);
        }))
    }
}

class postExecutor {
    constructor(private readonly url: string) {
    }

    async call(id: number, payload: string): Promise<object> {
        const fetchParams = {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: payload
        };
        const res = await fetch(this.url, fetchParams);
        if (!res.ok) {
            throw new Error(res.status + ' ' + res.statusText);
        }
        return await res.json();
    }
}

/**
API for managing queues
**/
export class QueuesAPI {

    private __id: number;
    private __executor:rpcExecutor;


    // Create new API handler to QueuesAPI.
    constructor(base_url : string = 'ws://127.0.0.1:3434/u/') {
        const proto = (new URL(base_url)).protocol;
        switch (proto) {
            case "ws:":
            case "wss:":{
                this.__executor=new wsExecutor(base_url);
                break
            }
            case "http:":
            case "https:":
            default:{
                this.__executor = new postExecutor(base_url);
                break
            }
        }
        this.__id = 1;
    }


    /**
    Create queue and link it to lambda and start worker
    **/
    async create(token: Token, queue: Queue): Promise<Queue> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Create",
            "id" : this.__next_id(),
            "params" : [token, queue]
        })) as Queue;
    }

    /**
    Remove queue and stop worker
    **/
    async remove(token: Token, name: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, name]
        })) as boolean;
    }

    /**
    Linked queues for lambda
    **/
    async linked(token: Token, lambda: string): Promise<Array<Queue>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Linked",
            "id" : this.__next_id(),
            "params" : [token, lambda]
        })) as Array<Queue>;
    }

    /**
    List of all queues
    **/
    async list(token: Token): Promise<Array<Queue>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Array<Queue>;
    }

    /**
    Assign lambda to queue (re-link)
    **/
    async assign(token: Token, name: string, lambda: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Assign",
            "id" : this.__next_id(),
            "params" : [token, name, lambda]
        })) as boolean;
    }


    private __next_id() {
        this.__id += 1;
        return this.__id
    }

    private async __call(req: { id: number, jsonrpc: string, method: string, params: object | Array<any> }): Promise<any> {
        const data = await this.__executor.call(req.id, JSON.stringify(req)) as {
            error?: {
                message: string,
                code: number,
                data?: any
            },
            result?:any
        }

        if (data.error) {
            throw new QueuesAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
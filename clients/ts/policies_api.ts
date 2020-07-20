export class PoliciesAPIError extends Error {
    public readonly code: number;
    public readonly details: any;

    constructor(message: string, code: number, details: any) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}


export interface Policy {
    id: string
    definition: PolicyDefinition
    lambdas: JsonStringSet
}

export interface PolicyDefinition {
    allowed_ip: JsonStringSet | null
    allowed_origin: JsonStringSet | null
    public: boolean
    tokens: any | null
}

export interface JsonStringSet {
}

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
API for managing policies
**/
export class PoliciesAPI {

    private __id: number;
    private __executor:rpcExecutor;


    // Create new API handler to PoliciesAPI.
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
    List all policies
    **/
    async list(token: Token): Promise<Array<Policy>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Array<Policy>;
    }

    /**
    Create new policy
    **/
    async create(token: Token, policy: string, definition: PolicyDefinition): Promise<Policy> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Create",
            "id" : this.__next_id(),
            "params" : [token, policy, definition]
        })) as Policy;
    }

    /**
    Remove policy
    **/
    async remove(token: Token, policy: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, policy]
        })) as boolean;
    }

    /**
    Update policy definition
    **/
    async update(token: Token, policy: string, definition: PolicyDefinition): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Update",
            "id" : this.__next_id(),
            "params" : [token, policy, definition]
        })) as boolean;
    }

    /**
    Apply policy for the resource
    **/
    async apply(token: Token, lambda: string, policy: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Apply",
            "id" : this.__next_id(),
            "params" : [token, lambda, policy]
        })) as boolean;
    }

    /**
    Clear applied policy for the lambda
    **/
    async clear(token: Token, lambda: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Clear",
            "id" : this.__next_id(),
            "params" : [token, lambda]
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
            throw new PoliciesAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
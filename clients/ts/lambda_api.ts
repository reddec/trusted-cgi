export class LambdaAPIError extends Error {
    public readonly code: number;
    public readonly details: any;

    constructor(message: string, code: number, details: any) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}


export type Token = string;

export interface File {
    name: string
    is_dir: boolean
}

export interface Definition {
    uid: string
    aliases: JsonStringSet
    manifest: Manifest
}

export interface JsonStringSet {
}

export interface Manifest {
    name: string | null
    description: string | null
    run: Array<string>
    output_headers: any | null
    input_headers: any | null
    query: any | null
    environment: any | null
    method: string | null
    method_env: string | null
    path_env: string | null
    time_limit: JsonDuration | null
    maximum_payload: number | null
    allowed_ip: JsonStringSet | null
    allowed_origin: JsonStringSet | null
    public: boolean
    tokens: any | null
    cron: Array<Schedule> | null
    static: string | null
}

export type JsonDuration = string; // suffixes: ns, us, ms, s, m, h

export interface Schedule {
    cron: string
    action: string
    time_limit: JsonDuration
}

export interface Record {
    uid: string
    error: string | null
    request: Request
    begin: Time
    end: Time
}

export interface Request {
    method: string
    url: string
    path: string
    remote_address: string
    form: any
    headers: any
}

export type Time = string; // RFC3339




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
API for lambdas
**/
export class LambdaAPI {

    private __id: number;
    private __executor:rpcExecutor;


    // Create new API handler to LambdaAPI.
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
    Upload content from .tar.gz archive to app and call Install handler (if defined)
    **/
    async upload(token: Token, uid: string, tarGz: Array<number>): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Upload",
            "id" : this.__next_id(),
            "params" : [token, uid, tarGz]
        })) as boolean;
    }

    /**
    Download content as .tar.gz archive from app
    **/
    async download(token: Token, uid: string): Promise<Array<number>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Download",
            "id" : this.__next_id(),
            "params" : [token, uid]
        })) as Array<number>;
    }

    /**
    Push single file to app
    **/
    async push(token: Token, uid: string, file: string, content: Array<number>): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Push",
            "id" : this.__next_id(),
            "params" : [token, uid, file, content]
        })) as boolean;
    }

    /**
    Pull single file from app
    **/
    async pull(token: Token, uid: string, file: string): Promise<Array<number>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Pull",
            "id" : this.__next_id(),
            "params" : [token, uid, file]
        })) as Array<number>;
    }

    /**
    Remove app and call Uninstall handler (if defined)
    **/
    async remove(token: Token, uid: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, uid]
        })) as boolean;
    }

    /**
    Files in func dir
    **/
    async files(token: Token, uid: string, dir: string): Promise<Array<File>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Files",
            "id" : this.__next_id(),
            "params" : [token, uid, dir]
        })) as Array<File>;
    }

    /**
    Info about application
    **/
    async info(token: Token, uid: string): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Info",
            "id" : this.__next_id(),
            "params" : [token, uid]
        })) as Definition;
    }

    /**
    Update application manifest
    **/
    async update(token: Token, uid: string, manifest: Manifest): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Update",
            "id" : this.__next_id(),
            "params" : [token, uid, manifest]
        })) as Definition;
    }

    /**
    Create file or directory inside app
    **/
    async createFile(token: Token, uid: string, path: string, dir: boolean): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.CreateFile",
            "id" : this.__next_id(),
            "params" : [token, uid, path, dir]
        })) as boolean;
    }

    /**
    Remove file or directory
    **/
    async removeFile(token: Token, uid: string, path: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.RemoveFile",
            "id" : this.__next_id(),
            "params" : [token, uid, path]
        })) as boolean;
    }

    /**
    Rename file or directory
    **/
    async renameFile(token: Token, uid: string, oldPath: string, newPath: string): Promise<boolean> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.RenameFile",
            "id" : this.__next_id(),
            "params" : [token, uid, oldPath, newPath]
        })) as boolean;
    }

    /**
    Stats for the app
    **/
    async stats(token: Token, uid: string, limit: number): Promise<Array<Record>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Stats",
            "id" : this.__next_id(),
            "params" : [token, uid, limit]
        })) as Array<Record>;
    }

    /**
    Actions available for the app
    **/
    async actions(token: Token, uid: string): Promise<Array<string>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Actions",
            "id" : this.__next_id(),
            "params" : [token, uid]
        })) as Array<string>;
    }

    /**
    Invoke action in the app (if make installed)
    **/
    async invoke(token: Token, uid: string, action: string): Promise<string> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Invoke",
            "id" : this.__next_id(),
            "params" : [token, uid, action]
        })) as string;
    }

    /**
    Make link/alias for app
    **/
    async link(token: Token, uid: string, alias: string): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Link",
            "id" : this.__next_id(),
            "params" : [token, uid, alias]
        })) as Definition;
    }

    /**
    Remove link
    **/
    async unlink(token: Token, alias: string): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Unlink",
            "id" : this.__next_id(),
            "params" : [token, alias]
        })) as Definition;
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
            throw new LambdaAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
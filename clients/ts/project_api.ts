export class ProjectAPIError extends Error {
    public readonly code: number;
    public readonly details: any;

    constructor(message: string, code: number, details: any) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}


export interface Settings {
    user: string
    public_key: string | null
    environment: any | null
}

export type Token = string;

export interface Environment {
    environment: any | null
}

export interface TemplateStatus {
    name: string
    description: string
    available: boolean
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

export interface Template {
    name: string
    description: string
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
API for global project
**/
export class ProjectAPI {

    private __id: number;
    private __executor:rpcExecutor;


    // Create new API handler to ProjectAPI.
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
    Get global configuration
    **/
    async config(token: Token): Promise<Settings> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Config",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Settings;
    }

    /**
    Change effective user
    **/
    async setUser(token: Token, user: string): Promise<Settings> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.SetUser",
            "id" : this.__next_id(),
            "params" : [token, user]
        })) as Settings;
    }

    /**
    Change global environment
    **/
    async setEnvironment(token: Token, env: Environment): Promise<Settings> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.SetEnvironment",
            "id" : this.__next_id(),
            "params" : [token, env]
        })) as Settings;
    }

    /**
    Get all templates without filtering
    **/
    async allTemplates(token: Token): Promise<Array<TemplateStatus>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.AllTemplates",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Array<TemplateStatus>;
    }

    /**
    List available apps (lambdas) in a project
    **/
    async list(token: Token): Promise<Array<Definition>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Array<Definition>;
    }

    /**
    Templates with filter by availability including embedded
    **/
    async templates(token: Token): Promise<Array<Template>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Templates",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Array<Template>;
    }

    /**
    Global last records
    **/
    async stats(token: Token, limit: number): Promise<Array<Record>> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Stats",
            "id" : this.__next_id(),
            "params" : [token, limit]
        })) as Array<Record>;
    }

    /**
    Create new app (lambda)
    **/
    async create(token: Token): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Create",
            "id" : this.__next_id(),
            "params" : [token]
        })) as Definition;
    }

    /**
    Create new app/lambda/function using pre-defined template
    **/
    async createFromTemplate(token: Token, templateName: string): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.CreateFromTemplate",
            "id" : this.__next_id(),
            "params" : [token, templateName]
        })) as Definition;
    }

    /**
    Create new app/lambda/function using remote Git repo
    **/
    async createFromGit(token: Token, repo: string): Promise<Definition> {
        return (await this.__call({
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.CreateFromGit",
            "id" : this.__next_id(),
            "params" : [token, repo]
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
            throw new ProjectAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
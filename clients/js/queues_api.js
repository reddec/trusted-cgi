export class QueuesAPIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class QueuesAPI {
    /**
    API for managing queues
    **/

    // Create new API handler to QueuesAPI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'https://127.0.0.1:3434/u/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Create queue and link it to lambda and start worker
    **/
    async create(token, name, lambda){
        return (await this.__call('Create', {
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Create",
            "id" : this.__next_id(),
            "params" : [token, name, lambda]
        }));
    }

    /**
    Remove queue and stop worker
    **/
    async remove(token, name){
        return (await this.__call('Remove', {
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, name]
        }));
    }

    /**
    Linked queues for lambda
    **/
    async linked(token, lambda){
        return (await this.__call('Linked', {
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Linked",
            "id" : this.__next_id(),
            "params" : [token, lambda]
        }));
    }

    /**
    List of all queues
    **/
    async list(token){
        return (await this.__call('List', {
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Assign lambda to queue (re-link)
    **/
    async assign(token, name, lambda){
        return (await this.__call('Assign', {
            "jsonrpc" : "2.0",
            "method" : "QueuesAPI.Assign",
            "id" : this.__next_id(),
            "params" : [token, name, lambda]
        }));
    }



    __next_id() {
        this.__id += 1;
        return this.__id
    }

    async __call(method, req) {
        const fetchParams = {
            method: "POST",
            headers: {
                'Content-Type' : 'application/json',
            },
            body: JSON.stringify(req)
        };
        if (this.__preflightHandler) {
            await Promise.resolve(this.__preflightHandler(method, fetchParams));
        }
        const res = await fetch(this.__url, fetchParams);
        if (!res.ok) {
            throw new Error(res.status + ' ' + res.statusText);
        }

        const data = await res.json();

        if ('error' in data) {
            throw new QueuesAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
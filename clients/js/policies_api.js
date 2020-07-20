export class PoliciesAPIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class PoliciesAPI {
    /**
    API for managing policies
    **/

    // Create new API handler to PoliciesAPI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'https://127.0.0.1:3434/u/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    List all policies
    **/
    async list(token){
        return (await this.__call('List', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Create new policy
    **/
    async create(token, policy, definition){
        return (await this.__call('Create', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Create",
            "id" : this.__next_id(),
            "params" : [token, policy, definition]
        }));
    }

    /**
    Remove policy
    **/
    async remove(token, policy){
        return (await this.__call('Remove', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, policy]
        }));
    }

    /**
    Update policy definition
    **/
    async update(token, policy, definition){
        return (await this.__call('Update', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Update",
            "id" : this.__next_id(),
            "params" : [token, policy, definition]
        }));
    }

    /**
    Apply policy for the resource
    **/
    async apply(token, lambda, policy){
        return (await this.__call('Apply', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Apply",
            "id" : this.__next_id(),
            "params" : [token, lambda, policy]
        }));
    }

    /**
    Clear applied policy for the lambda
    **/
    async clear(token, lambda){
        return (await this.__call('Clear', {
            "jsonrpc" : "2.0",
            "method" : "PoliciesAPI.Clear",
            "id" : this.__next_id(),
            "params" : [token, lambda]
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
            throw new PoliciesAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
export class ProjectAPIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class ProjectAPI {
    /**
    Remove link
    **/

    // Create new API handler to ProjectAPI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'https://127.0.0.1:3434/u/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Get global configuration
    **/
    async config(token){
        return (await this.__call('Config', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Config",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Change effective user
    **/
    async setUser(token, user){
        return (await this.__call('SetUser', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.SetUser",
            "id" : this.__next_id(),
            "params" : [token, user]
        }));
    }

    /**
    Get all templates without filtering
    **/
    async allTemplates(token){
        return (await this.__call('AllTemplates', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.AllTemplates",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    List available apps (lambdas) in a project
    **/
    async list(token){
        return (await this.__call('List', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.List",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Templates with filter by availability including embedded
    **/
    async templates(token){
        return (await this.__call('Templates', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Templates",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Global last records
    **/
    async stats(token, limit){
        return (await this.__call('Stats', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Stats",
            "id" : this.__next_id(),
            "params" : [token, limit]
        }));
    }

    /**
    Create new app (lambda)
    **/
    async create(token){
        return (await this.__call('Create', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.Create",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Create new app/lambda/function using pre-defined template
    **/
    async createFromTemplate(token, templateName){
        return (await this.__call('CreateFromTemplate', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.CreateFromTemplate",
            "id" : this.__next_id(),
            "params" : [token, templateName]
        }));
    }

    /**
    Create new app/lambda/function using remote Git repo
    **/
    async createFromGit(token, repo){
        return (await this.__call('CreateFromGit', {
            "jsonrpc" : "2.0",
            "method" : "ProjectAPI.CreateFromGit",
            "id" : this.__next_id(),
            "params" : [token, repo]
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
            throw new ProjectAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
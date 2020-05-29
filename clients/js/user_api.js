export class UserAPIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class UserAPI {
    /**
    User/admin profile API
    **/

    // Create new API handler to UserAPI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'https://127.0.0.1:3434/u/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Login user by username and password. Returns signed JWT
    **/
    async login(login, password){
        return (await this.__call('Login', {
            "jsonrpc" : "2.0",
            "method" : "UserAPI.Login",
            "id" : this.__next_id(),
            "params" : [login, password]
        }));
    }

    /**
    Change password for the user
    **/
    async changePassword(token, password){
        return (await this.__call('ChangePassword', {
            "jsonrpc" : "2.0",
            "method" : "UserAPI.ChangePassword",
            "id" : this.__next_id(),
            "params" : [token, password]
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
            throw new UserAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
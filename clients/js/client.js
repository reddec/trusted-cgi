export class APIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class API {
    /**
    
    **/

    // Create new API handler to API.
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
            "method" : "API.Login",
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
            "method" : "API.ChangePassword",
            "id" : this.__next_id(),
            "params" : [token, password]
        }));
    }

    /**
    Create new app (lambda)
    **/
    async create(token){
        return (await this.__call('Create', {
            "jsonrpc" : "2.0",
            "method" : "API.Create",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Project configuration
    **/
    async config(token){
        return (await this.__call('Config', {
            "jsonrpc" : "2.0",
            "method" : "API.Config",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Apply new configuration and save it
    **/
    async apply(token, config){
        return (await this.__call('Apply', {
            "jsonrpc" : "2.0",
            "method" : "API.Apply",
            "id" : this.__next_id(),
            "params" : [token, config]
        }));
    }

    /**
    Get all templates without filtering
    **/
    async allTemplates(token){
        return (await this.__call('AllTemplates', {
            "jsonrpc" : "2.0",
            "method" : "API.AllTemplates",
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
            "method" : "API.CreateFromTemplate",
            "id" : this.__next_id(),
            "params" : [token, templateName]
        }));
    }

    /**
    Upload content from .tar.gz archive to app and call Install handler (if defined)
    **/
    async upload(token, uid, tarGz){
        return (await this.__call('Upload', {
            "jsonrpc" : "2.0",
            "method" : "API.Upload",
            "id" : this.__next_id(),
            "params" : [token, uid, tarGz]
        }));
    }

    /**
    Download content as .tar.gz archive from app
    **/
    async download(token, uid){
        return (await this.__call('Download', {
            "jsonrpc" : "2.0",
            "method" : "API.Download",
            "id" : this.__next_id(),
            "params" : [token, uid]
        }));
    }

    /**
    Push single file to app
    **/
    async push(token, uid, file, content){
        return (await this.__call('Push', {
            "jsonrpc" : "2.0",
            "method" : "API.Push",
            "id" : this.__next_id(),
            "params" : [token, uid, file, content]
        }));
    }

    /**
    Pull single file from app
    **/
    async pull(token, uid, file){
        return (await this.__call('Pull', {
            "jsonrpc" : "2.0",
            "method" : "API.Pull",
            "id" : this.__next_id(),
            "params" : [token, uid, file]
        }));
    }

    /**
    List available apps (lambdas) in a project
    **/
    async list(token){
        return (await this.__call('List', {
            "jsonrpc" : "2.0",
            "method" : "API.List",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Remove app and call Uninstall handler (if defined)
    **/
    async remove(token, uid){
        return (await this.__call('Remove', {
            "jsonrpc" : "2.0",
            "method" : "API.Remove",
            "id" : this.__next_id(),
            "params" : [token, uid]
        }));
    }

    /**
    Templates with filter by availability including embedded
    **/
    async templates(token){
        return (await this.__call('Templates', {
            "jsonrpc" : "2.0",
            "method" : "API.Templates",
            "id" : this.__next_id(),
            "params" : [token]
        }));
    }

    /**
    Files in func dir
    **/
    async files(token, name, dir){
        return (await this.__call('Files', {
            "jsonrpc" : "2.0",
            "method" : "API.Files",
            "id" : this.__next_id(),
            "params" : [token, name, dir]
        }));
    }

    /**
    Info about application
    **/
    async info(token, uid){
        return (await this.__call('Info', {
            "jsonrpc" : "2.0",
            "method" : "API.Info",
            "id" : this.__next_id(),
            "params" : [token, uid]
        }));
    }

    /**
    Update application manifest
    **/
    async update(token, uid, manifest){
        return (await this.__call('Update', {
            "jsonrpc" : "2.0",
            "method" : "API.Update",
            "id" : this.__next_id(),
            "params" : [token, uid, manifest]
        }));
    }

    /**
    Create file or directory inside app
    **/
    async createFile(token, uid, path, dir){
        return (await this.__call('CreateFile', {
            "jsonrpc" : "2.0",
            "method" : "API.CreateFile",
            "id" : this.__next_id(),
            "params" : [token, uid, path, dir]
        }));
    }

    /**
    Remove file or directory
    **/
    async removeFile(token, uid, path){
        return (await this.__call('RemoveFile', {
            "jsonrpc" : "2.0",
            "method" : "API.RemoveFile",
            "id" : this.__next_id(),
            "params" : [token, uid, path]
        }));
    }

    /**
    Rename file or directory
    **/
    async renameFile(token, uid, oldPath, newPath){
        return (await this.__call('RenameFile', {
            "jsonrpc" : "2.0",
            "method" : "API.RenameFile",
            "id" : this.__next_id(),
            "params" : [token, uid, oldPath, newPath]
        }));
    }

    /**
    Global last records
    **/
    async globalStats(token, limit){
        return (await this.__call('GlobalStats', {
            "jsonrpc" : "2.0",
            "method" : "API.GlobalStats",
            "id" : this.__next_id(),
            "params" : [token, limit]
        }));
    }

    /**
    Stats
    **/
    async stats(token, uid, limit){
        return (await this.__call('Stats', {
            "jsonrpc" : "2.0",
            "method" : "API.Stats",
            "id" : this.__next_id(),
            "params" : [token, uid, limit]
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
            throw new APIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
export class LambdaAPIError extends Error {
    constructor(message, code, details) {
        super(code + ': ' + message);
        this.code = code;
        this.details = details;
    }
}

export class LambdaAPI {
    /**
    optional public RSA key for SSH
    **/

    // Create new API handler to LambdaAPI.
    // preflightHandler (if defined) can return promise
    constructor(base_url = 'https://127.0.0.1:3434/u/', preflightHandler = null) {
        this.__url = base_url;
        this.__id = 1;
        this.__preflightHandler = preflightHandler;
    }


    /**
    Upload content from .tar.gz archive to app and call Install handler (if defined)
    **/
    async upload(token, uid, tarGz){
        return (await this.__call('Upload', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Upload",
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
            "method" : "LambdaAPI.Download",
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
            "method" : "LambdaAPI.Push",
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
            "method" : "LambdaAPI.Pull",
            "id" : this.__next_id(),
            "params" : [token, uid, file]
        }));
    }

    /**
    Remove app and call Uninstall handler (if defined)
    **/
    async remove(token, uid){
        return (await this.__call('Remove', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Remove",
            "id" : this.__next_id(),
            "params" : [token, uid]
        }));
    }

    /**
    Files in func dir
    **/
    async files(token, uid, dir){
        return (await this.__call('Files', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Files",
            "id" : this.__next_id(),
            "params" : [token, uid, dir]
        }));
    }

    /**
    Info about application
    **/
    async info(token, uid){
        return (await this.__call('Info', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Info",
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
            "method" : "LambdaAPI.Update",
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
            "method" : "LambdaAPI.CreateFile",
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
            "method" : "LambdaAPI.RemoveFile",
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
            "method" : "LambdaAPI.RenameFile",
            "id" : this.__next_id(),
            "params" : [token, uid, oldPath, newPath]
        }));
    }

    /**
    Stats for the app
    **/
    async stats(token, uid, limit){
        return (await this.__call('Stats', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Stats",
            "id" : this.__next_id(),
            "params" : [token, uid, limit]
        }));
    }

    /**
    Actions available for the app
    **/
    async actions(token, uid){
        return (await this.__call('Actions', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Actions",
            "id" : this.__next_id(),
            "params" : [token, uid]
        }));
    }

    /**
    Invoke action in the app (if make installed)
    **/
    async invoke(token, uid, action){
        return (await this.__call('Invoke', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Invoke",
            "id" : this.__next_id(),
            "params" : [token, uid, action]
        }));
    }

    /**
    Make link/alias for app
    **/
    async link(token, uid, alias){
        return (await this.__call('Link', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Link",
            "id" : this.__next_id(),
            "params" : [token, uid, alias]
        }));
    }

    /**
    Remove link
    **/
    async unlink(token, alias){
        return (await this.__call('Unlink', {
            "jsonrpc" : "2.0",
            "method" : "LambdaAPI.Unlink",
            "id" : this.__next_id(),
            "params" : [token, alias]
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
            throw new LambdaAPIError(data.error.message, data.error.code, data.error.data);
        }

        return data.result;
    }
}
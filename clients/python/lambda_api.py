from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional
from base64 import decodebytes, encodebytes



@dataclass
class File:
    dir: 'bool'
    name: 'str'

    def to_json(self) -> dict:
        return {
            "is_dir": self.dir,
            "name": self.name,
        }

    @staticmethod
    def from_json(payload: dict) -> 'File':
        return File(
                dir=payload['is_dir'],
                name=payload['name'],
        )


@dataclass
class App:
    uid: 'str'
    manifest: 'Manifest'
    is_git: 'bool'

    def to_json(self) -> dict:
        return {
            "uid": self.uid,
            "manifest": self.manifest.to_json(),
            "git": self.is_git,
        }

    @staticmethod
    def from_json(payload: dict) -> 'App':
        return App(
                uid=payload['uid'],
                manifest=Manifest.from_json(payload['manifest']),
                is_git=payload['git'],
        )


@dataclass
class Manifest:
    name: 'str'
    description: 'str'
    run: 'List[str]'
    output_headers: 'Any'
    input_headers: 'Optional[Any]'
    query: 'Optional[Any]'
    environment: 'Optional[Any]'
    method: 'Optional[str]'
    method_env: 'Optional[str]'
    path_env: 'Optional[str]'
    time_limit: 'Optional[Any]'
    maximum_payload: 'Optional[int]'
    allowed_ip: 'Optional[Any]'
    allowed_origin: 'Optional[Any]'
    public: 'bool'
    tokens: 'Optional[Any]'
    post_clone: 'Optional[str]'
    aliases: 'Optional[Any]'
    cron: 'Optional[List[Schedule]]'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "description": self.description,
            "run": self.run,
            "output_headers": self.output_headers,
            "input_headers": self.input_headers,
            "query": self.query,
            "environment": self.environment,
            "method": self.method,
            "method_env": self.method_env,
            "path_env": self.path_env,
            "time_limit": self.time_limit,
            "maximum_payload": self.maximum_payload,
            "allowed_ip": self.allowed_ip,
            "allowed_origin": self.allowed_origin,
            "public": self.public,
            "tokens": self.tokens,
            "post_clone": self.post_clone,
            "aliases": self.aliases,
            "cron": [x.to_json() for x in self.cron],
        }

    @staticmethod
    def from_json(payload: dict) -> 'Manifest':
        return Manifest(
                name=payload['name'],
                description=payload['description'],
                run=payload['run'] or [],
                output_headers=payload['output_headers'],
                input_headers=payload['input_headers'],
                query=payload['query'],
                environment=payload['environment'],
                method=payload['method'],
                method_env=payload['method_env'],
                path_env=payload['path_env'],
                time_limit=payload['time_limit'],
                maximum_payload=payload['maximum_payload'],
                allowed_ip=payload['allowed_ip'],
                allowed_origin=payload['allowed_origin'],
                public=payload['public'],
                tokens=payload['tokens'],
                post_clone=payload['post_clone'],
                aliases=payload['aliases'],
                cron=[Schedule.from_json(x) for x in (payload['cron'] or [])],
        )


@dataclass
class Schedule:
    cron: 'str'
    action: 'str'
    time_limit: 'Any'

    def to_json(self) -> dict:
        return {
            "cron": self.cron,
            "action": self.action,
            "time_limit": self.time_limit,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Schedule':
        return Schedule(
                cron=payload['cron'],
                action=payload['action'],
                time_limit=payload['time_limit'],
        )


@dataclass
class Record:
    uid: 'str'
    input: 'Optional[bytes]'
    output: 'Optional[bytes]'
    err: 'Optional[str]'
    code: 'int'
    method: 'str'
    remote: 'str'
    origin: 'Optional[str]'
    uri: 'str'
    token: 'Optional[str]'
    begin: 'Any'
    end: 'Any'

    def to_json(self) -> dict:
        return {
            "uid": self.uid,
            "input": encodebytes(self.input),
            "output": encodebytes(self.output),
            "error": self.err,
            "code": self.code,
            "method": self.method,
            "remote": self.remote,
            "origin": self.origin,
            "uri": self.uri,
            "token": self.token,
            "begin": self.begin,
            "end": self.end,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Record':
        return Record(
                uid=payload['uid'],
                input=decodebytes((payload['input'] or '').encode()),
                output=decodebytes((payload['output'] or '').encode()),
                err=payload['error'],
                code=payload['code'],
                method=payload['method'],
                remote=payload['remote'],
                origin=payload['origin'],
                uri=payload['uri'],
                token=payload['token'],
                begin=payload['begin'],
                end=payload['end'],
        )


class LambdaAPIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'LambdaAPIError':
        return LambdaAPIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class LambdaAPIClient:
    """
    API for lambdas
    """

    def __init__(self, base_url: str = 'https://127.0.0.1:3434/u/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def upload(self, token: Any, uid: str, tar_gz: bytes) -> bool:
        """
        Upload content from .tar.gz archive to app and call Install handler (if defined)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Upload",
            "id": self.__next_id(),
            "params": [token, uid, encodebytes(tar_gz), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('upload', payload['error'])
        return payload['result']

    async def download(self, token: Any, uid: str) -> bytes:
        """
        Download content as .tar.gz archive from app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Download",
            "id": self.__next_id(),
            "params": [token, uid, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('download', payload['error'])
        return decodebytes((payload['result'] or '').encode())

    async def push(self, token: Any, uid: str, file: str, content: bytes) -> bool:
        """
        Push single file to app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Push",
            "id": self.__next_id(),
            "params": [token, uid, file, encodebytes(content), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('push', payload['error'])
        return payload['result']

    async def pull(self, token: Any, uid: str, file: str) -> bytes:
        """
        Pull single file from app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Pull",
            "id": self.__next_id(),
            "params": [token, uid, file, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('pull', payload['error'])
        return decodebytes((payload['result'] or '').encode())

    async def remove(self, token: Any, uid: str) -> bool:
        """
        Remove app and call Uninstall handler (if defined)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Remove",
            "id": self.__next_id(),
            "params": [token, uid, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('remove', payload['error'])
        return payload['result']

    async def files(self, token: Any, uid: str, dir: str) -> List[File]:
        """
        Files in func dir
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Files",
            "id": self.__next_id(),
            "params": [token, uid, dir, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('files', payload['error'])
        return [File.from_json(x) for x in (payload['result'] or [])]

    async def info(self, token: Any, uid: str) -> App:
        """
        Info about application
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Info",
            "id": self.__next_id(),
            "params": [token, uid, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('info', payload['error'])
        return App.from_json(payload['result'])

    async def update(self, token: Any, uid: str, manifest: Manifest) -> App:
        """
        Update application manifest
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Update",
            "id": self.__next_id(),
            "params": [token, uid, manifest.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('update', payload['error'])
        return App.from_json(payload['result'])

    async def create_file(self, token: Any, uid: str, path: str, dir: bool) -> bool:
        """
        Create file or directory inside app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.CreateFile",
            "id": self.__next_id(),
            "params": [token, uid, path, dir, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('create_file', payload['error'])
        return payload['result']

    async def remove_file(self, token: Any, uid: str, path: str) -> bool:
        """
        Remove file or directory
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.RemoveFile",
            "id": self.__next_id(),
            "params": [token, uid, path, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('remove_file', payload['error'])
        return payload['result']

    async def rename_file(self, token: Any, uid: str, old_path: str, new_path: str) -> bool:
        """
        Rename file or directory
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.RenameFile",
            "id": self.__next_id(),
            "params": [token, uid, old_path, new_path, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('rename_file', payload['error'])
        return payload['result']

    async def stats(self, token: Any, uid: str, limit: int) -> List[Record]:
        """
        Stats for the app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Stats",
            "id": self.__next_id(),
            "params": [token, uid, limit, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('stats', payload['error'])
        return [Record.from_json(x) for x in (payload['result'] or [])]

    async def actions(self, token: Any, uid: str) -> List[str]:
        """
        Actions available for the app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Actions",
            "id": self.__next_id(),
            "params": [token, uid, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('actions', payload['error'])
        return payload['result'] or []

    async def invoke(self, token: Any, uid: str, action: str) -> str:
        """
        Invoke action in the app (if make installed)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Invoke",
            "id": self.__next_id(),
            "params": [token, uid, action, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('invoke', payload['error'])
        return payload['result']

    async def link(self, token: Any, uid: str, alias: str) -> App:
        """
        Make link/alias for app
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Link",
            "id": self.__next_id(),
            "params": [token, uid, alias, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('link', payload['error'])
        return App.from_json(payload['result'])

    async def unlink(self, token: Any, alias: str) -> App:
        """
        Remove link
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "LambdaAPI.Unlink",
            "id": self.__next_id(),
            "params": [token, alias, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise LambdaAPIError.from_json('unlink', payload['error'])
        return App.from_json(payload['result'])

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class LambdaAPIBatch:
    """
    API for lambdas
    """

    def __init__(self, client: LambdaAPIClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def upload(self, token: Any, uid: str, tar_gz: bytes):
        """
        Upload content from .tar.gz archive to app and call Install handler (if defined)
        """
        params = [token, uid, encodebytes(tar_gz), ]
        method = "LambdaAPI.Upload"
        self.__add_request(method, params, lambda payload: payload)

    def download(self, token: Any, uid: str):
        """
        Download content as .tar.gz archive from app
        """
        params = [token, uid, ]
        method = "LambdaAPI.Download"
        self.__add_request(method, params, lambda payload: decodebytes((payload or '').encode()))

    def push(self, token: Any, uid: str, file: str, content: bytes):
        """
        Push single file to app
        """
        params = [token, uid, file, encodebytes(content), ]
        method = "LambdaAPI.Push"
        self.__add_request(method, params, lambda payload: payload)

    def pull(self, token: Any, uid: str, file: str):
        """
        Pull single file from app
        """
        params = [token, uid, file, ]
        method = "LambdaAPI.Pull"
        self.__add_request(method, params, lambda payload: decodebytes((payload or '').encode()))

    def remove(self, token: Any, uid: str):
        """
        Remove app and call Uninstall handler (if defined)
        """
        params = [token, uid, ]
        method = "LambdaAPI.Remove"
        self.__add_request(method, params, lambda payload: payload)

    def files(self, token: Any, uid: str, dir: str):
        """
        Files in func dir
        """
        params = [token, uid, dir, ]
        method = "LambdaAPI.Files"
        self.__add_request(method, params, lambda payload: [File.from_json(x) for x in (payload or [])])

    def info(self, token: Any, uid: str):
        """
        Info about application
        """
        params = [token, uid, ]
        method = "LambdaAPI.Info"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def update(self, token: Any, uid: str, manifest: Manifest):
        """
        Update application manifest
        """
        params = [token, uid, manifest.to_json(), ]
        method = "LambdaAPI.Update"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def create_file(self, token: Any, uid: str, path: str, dir: bool):
        """
        Create file or directory inside app
        """
        params = [token, uid, path, dir, ]
        method = "LambdaAPI.CreateFile"
        self.__add_request(method, params, lambda payload: payload)

    def remove_file(self, token: Any, uid: str, path: str):
        """
        Remove file or directory
        """
        params = [token, uid, path, ]
        method = "LambdaAPI.RemoveFile"
        self.__add_request(method, params, lambda payload: payload)

    def rename_file(self, token: Any, uid: str, old_path: str, new_path: str):
        """
        Rename file or directory
        """
        params = [token, uid, old_path, new_path, ]
        method = "LambdaAPI.RenameFile"
        self.__add_request(method, params, lambda payload: payload)

    def stats(self, token: Any, uid: str, limit: int):
        """
        Stats for the app
        """
        params = [token, uid, limit, ]
        method = "LambdaAPI.Stats"
        self.__add_request(method, params, lambda payload: [Record.from_json(x) for x in (payload or [])])

    def actions(self, token: Any, uid: str):
        """
        Actions available for the app
        """
        params = [token, uid, ]
        method = "LambdaAPI.Actions"
        self.__add_request(method, params, lambda payload: payload or [])

    def invoke(self, token: Any, uid: str, action: str):
        """
        Invoke action in the app (if make installed)
        """
        params = [token, uid, action, ]
        method = "LambdaAPI.Invoke"
        self.__add_request(method, params, lambda payload: payload)

    def link(self, token: Any, uid: str, alias: str):
        """
        Make link/alias for app
        """
        params = [token, uid, alias, ]
        method = "LambdaAPI.Link"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def unlink(self, token: Any, alias: str):
        """
        Remove link
        """
        params = [token, alias, ]
        method = "LambdaAPI.Unlink"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def __add_request(self, method: str, params, factory):
        request_id = self.__next_id()
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": request_id,
            "params": params
        }
        self.__requests.append(request)
        self.__batch[request_id] = (request, factory)

    async def __aenter__(self):
        self.__batch = {}
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self()

    async def __call__(self) -> list:
        offset = 0
        num = len(self.__requests)
        results = []
        while offset < num:
            next_offset = offset + self.__batch_size
            batch = self.__requests[offset:min(num, next_offset)]
            offset = next_offset

            responses = await self.__post_batch(batch)
            results = results + responses

        self.__batch = {}
        self.__requests = []
        return results

    async def __post_batch(self, batch: list) -> list:
        response = await self.__client._invoke(batch)
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        results = await response.json()
        ans = []
        for payload in results:
            request, factory = self.__batch[payload['id']]
            if 'error' in payload:
                raise LambdaAPIError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans
from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional
from base64 import decodebytes, encodebytes



@dataclass
class Settings:
    user: 'str'
    public_key: 'Optional[str]'
    environment: 'Optional[Any]'

    def to_json(self) -> dict:
        return {
            "user": self.user,
            "public_key": self.public_key,
            "environment": self.environment,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Settings':
        return Settings(
                user=payload['user'],
                public_key=payload['public_key'],
                environment=payload['environment'],
        )


@dataclass
class Environment:
    environment: 'Optional[Any]'

    def to_json(self) -> dict:
        return {
            "environment": self.environment,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Environment':
        return Environment(
                environment=payload['environment'],
        )


@dataclass
class TemplateStatus:
    name: 'str'
    description: 'str'
    available: 'bool'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "description": self.description,
            "available": self.available,
        }

    @staticmethod
    def from_json(payload: dict) -> 'TemplateStatus':
        return TemplateStatus(
                name=payload['name'],
                description=payload['description'],
                available=payload['available'],
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
class Template:
    name: 'str'
    description: 'str'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "description": self.description,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Template':
        return Template(
                name=payload['name'],
                description=payload['description'],
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


class ProjectAPIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'ProjectAPIError':
        return ProjectAPIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class ProjectAPIClient:
    """
    API for global project
    """

    def __init__(self, base_url: str = 'https://127.0.0.1:3434/u/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def config(self, token: Any) -> Settings:
        """
        Get global configuration
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.Config",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('config', payload['error'])
        return Settings.from_json(payload['result'])

    async def set_user(self, token: Any, user: str) -> Settings:
        """
        Change effective user
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.SetUser",
            "id": self.__next_id(),
            "params": [token, user, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('set_user', payload['error'])
        return Settings.from_json(payload['result'])

    async def set_environment(self, token: Any, env: Environment) -> Settings:
        """
        Change global environment
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.SetEnvironment",
            "id": self.__next_id(),
            "params": [token, env.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('set_environment', payload['error'])
        return Settings.from_json(payload['result'])

    async def all_templates(self, token: Any) -> List[TemplateStatus]:
        """
        Get all templates without filtering
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.AllTemplates",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('all_templates', payload['error'])
        return [TemplateStatus.from_json(x) for x in (payload['result'] or [])]

    async def list(self, token: Any) -> List[App]:
        """
        List available apps (lambdas) in a project
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.List",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('list', payload['error'])
        return [App.from_json(x) for x in (payload['result'] or [])]

    async def templates(self, token: Any) -> List[Template]:
        """
        Templates with filter by availability including embedded
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.Templates",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('templates', payload['error'])
        return [Template.from_json(x) for x in (payload['result'] or [])]

    async def stats(self, token: Any, limit: int) -> List[Record]:
        """
        Global last records
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.Stats",
            "id": self.__next_id(),
            "params": [token, limit, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('stats', payload['error'])
        return [Record.from_json(x) for x in (payload['result'] or [])]

    async def create(self, token: Any) -> App:
        """
        Create new app (lambda)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.Create",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('create', payload['error'])
        return App.from_json(payload['result'])

    async def create_from_template(self, token: Any, template_name: str) -> App:
        """
        Create new app/lambda/function using pre-defined template
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.CreateFromTemplate",
            "id": self.__next_id(),
            "params": [token, template_name, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('create_from_template', payload['error'])
        return App.from_json(payload['result'])

    async def create_from_git(self, token: Any, repo: str) -> App:
        """
        Create new app/lambda/function using remote Git repo
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "ProjectAPI.CreateFromGit",
            "id": self.__next_id(),
            "params": [token, repo, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise ProjectAPIError.from_json('create_from_git', payload['error'])
        return App.from_json(payload['result'])

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class ProjectAPIBatch:
    """
    API for global project
    """

    def __init__(self, client: ProjectAPIClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def config(self, token: Any):
        """
        Get global configuration
        """
        params = [token, ]
        method = "ProjectAPI.Config"
        self.__add_request(method, params, lambda payload: Settings.from_json(payload))

    def set_user(self, token: Any, user: str):
        """
        Change effective user
        """
        params = [token, user, ]
        method = "ProjectAPI.SetUser"
        self.__add_request(method, params, lambda payload: Settings.from_json(payload))

    def set_environment(self, token: Any, env: Environment):
        """
        Change global environment
        """
        params = [token, env.to_json(), ]
        method = "ProjectAPI.SetEnvironment"
        self.__add_request(method, params, lambda payload: Settings.from_json(payload))

    def all_templates(self, token: Any):
        """
        Get all templates without filtering
        """
        params = [token, ]
        method = "ProjectAPI.AllTemplates"
        self.__add_request(method, params, lambda payload: [TemplateStatus.from_json(x) for x in (payload or [])])

    def list(self, token: Any):
        """
        List available apps (lambdas) in a project
        """
        params = [token, ]
        method = "ProjectAPI.List"
        self.__add_request(method, params, lambda payload: [App.from_json(x) for x in (payload or [])])

    def templates(self, token: Any):
        """
        Templates with filter by availability including embedded
        """
        params = [token, ]
        method = "ProjectAPI.Templates"
        self.__add_request(method, params, lambda payload: [Template.from_json(x) for x in (payload or [])])

    def stats(self, token: Any, limit: int):
        """
        Global last records
        """
        params = [token, limit, ]
        method = "ProjectAPI.Stats"
        self.__add_request(method, params, lambda payload: [Record.from_json(x) for x in (payload or [])])

    def create(self, token: Any):
        """
        Create new app (lambda)
        """
        params = [token, ]
        method = "ProjectAPI.Create"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def create_from_template(self, token: Any, template_name: str):
        """
        Create new app/lambda/function using pre-defined template
        """
        params = [token, template_name, ]
        method = "ProjectAPI.CreateFromTemplate"
        self.__add_request(method, params, lambda payload: App.from_json(payload))

    def create_from_git(self, token: Any, repo: str):
        """
        Create new app/lambda/function using remote Git repo
        """
        params = [token, repo, ]
        method = "ProjectAPI.CreateFromGit"
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
                raise ProjectAPIError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans
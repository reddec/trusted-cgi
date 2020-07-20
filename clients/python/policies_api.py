from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional



@dataclass
class Policy:
    id: 'str'
    definition: 'PolicyDefinition'
    lambdas: 'Any'

    def to_json(self) -> dict:
        return {
            "id": self.id,
            "definition": self.definition.to_json(),
            "lambdas": self.lambdas,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Policy':
        return Policy(
                id=payload['id'],
                definition=PolicyDefinition.from_json(payload['definition']),
                lambdas=payload['lambdas'],
        )


@dataclass
class PolicyDefinition:
    allowed_ip: 'Optional[Any]'
    allowed_origin: 'Optional[Any]'
    public: 'bool'
    tokens: 'Optional[Any]'

    def to_json(self) -> dict:
        return {
            "allowed_ip": self.allowed_ip,
            "allowed_origin": self.allowed_origin,
            "public": self.public,
            "tokens": self.tokens,
        }

    @staticmethod
    def from_json(payload: dict) -> 'PolicyDefinition':
        return PolicyDefinition(
                allowed_ip=payload['allowed_ip'],
                allowed_origin=payload['allowed_origin'],
                public=payload['public'],
                tokens=payload['tokens'],
        )


class PoliciesAPIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'PoliciesAPIError':
        return PoliciesAPIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class PoliciesAPIClient:
    """
    API for managing policies
    """

    def __init__(self, base_url: str = 'https://127.0.0.1:3434/u/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def list(self, token: Any) -> List[Policy]:
        """
        List all policies
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.List",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('list', payload['error'])
        return [Policy.from_json(x) for x in (payload['result'] or [])]

    async def create(self, token: Any, policy: str, definition: PolicyDefinition) -> Policy:
        """
        Create new policy
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.Create",
            "id": self.__next_id(),
            "params": [token, policy, definition.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('create', payload['error'])
        return Policy.from_json(payload['result'])

    async def remove(self, token: Any, policy: str) -> bool:
        """
        Remove policy
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.Remove",
            "id": self.__next_id(),
            "params": [token, policy, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('remove', payload['error'])
        return payload['result']

    async def update(self, token: Any, policy: str, definition: PolicyDefinition) -> bool:
        """
        Update policy definition
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.Update",
            "id": self.__next_id(),
            "params": [token, policy, definition.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('update', payload['error'])
        return payload['result']

    async def apply(self, token: Any, lambda: str, policy: str) -> bool:
        """
        Apply policy for the resource
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.Apply",
            "id": self.__next_id(),
            "params": [token, lambda, policy, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('apply', payload['error'])
        return payload['result']

    async def clear(self, token: Any, lambda: str) -> bool:
        """
        Clear applied policy for the lambda
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "PoliciesAPI.Clear",
            "id": self.__next_id(),
            "params": [token, lambda, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise PoliciesAPIError.from_json('clear', payload['error'])
        return payload['result']

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class PoliciesAPIBatch:
    """
    API for managing policies
    """

    def __init__(self, client: PoliciesAPIClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def list(self, token: Any):
        """
        List all policies
        """
        params = [token, ]
        method = "PoliciesAPI.List"
        self.__add_request(method, params, lambda payload: [Policy.from_json(x) for x in (payload or [])])

    def create(self, token: Any, policy: str, definition: PolicyDefinition):
        """
        Create new policy
        """
        params = [token, policy, definition.to_json(), ]
        method = "PoliciesAPI.Create"
        self.__add_request(method, params, lambda payload: Policy.from_json(payload))

    def remove(self, token: Any, policy: str):
        """
        Remove policy
        """
        params = [token, policy, ]
        method = "PoliciesAPI.Remove"
        self.__add_request(method, params, lambda payload: payload)

    def update(self, token: Any, policy: str, definition: PolicyDefinition):
        """
        Update policy definition
        """
        params = [token, policy, definition.to_json(), ]
        method = "PoliciesAPI.Update"
        self.__add_request(method, params, lambda payload: payload)

    def apply(self, token: Any, lambda: str, policy: str):
        """
        Apply policy for the resource
        """
        params = [token, lambda, policy, ]
        method = "PoliciesAPI.Apply"
        self.__add_request(method, params, lambda payload: payload)

    def clear(self, token: Any, lambda: str):
        """
        Clear applied policy for the lambda
        """
        params = [token, lambda, ]
        method = "PoliciesAPI.Clear"
        self.__add_request(method, params, lambda payload: payload)

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
                raise PoliciesAPIError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans
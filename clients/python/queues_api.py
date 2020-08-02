from aiohttp import client

from dataclasses import dataclass

from typing import Any, List, Optional



@dataclass
class Queue:
    name: 'str'
    target: 'str'
    retry: 'int'
    max_element_size: 'int'
    interval: 'Any'

    def to_json(self) -> dict:
        return {
            "name": self.name,
            "target": self.target,
            "retry": self.retry,
            "max_element_size": self.max_element_size,
            "interval": self.interval,
        }

    @staticmethod
    def from_json(payload: dict) -> 'Queue':
        return Queue(
                name=payload['name'],
                target=payload['target'],
                retry=payload['retry'],
                max_element_size=payload['max_element_size'],
                interval=payload['interval'],
        )


class QueuesAPIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'QueuesAPIError':
        return QueuesAPIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class QueuesAPIClient:
    """
    API for managing queues
    """

    def __init__(self, base_url: str = 'https://127.0.0.1:3434/u/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def create(self, token: Any, queue: Queue) -> Queue:
        """
        Create queue and link it to lambda and start worker
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "QueuesAPI.Create",
            "id": self.__next_id(),
            "params": [token, queue.to_json(), ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise QueuesAPIError.from_json('create', payload['error'])
        return Queue.from_json(payload['result'])

    async def remove(self, token: Any, name: str) -> bool:
        """
        Remove queue and stop worker
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "QueuesAPI.Remove",
            "id": self.__next_id(),
            "params": [token, name, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise QueuesAPIError.from_json('remove', payload['error'])
        return payload['result']

    async def linked(self, token: Any, lambda: str) -> List[Queue]:
        """
        Linked queues for lambda
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "QueuesAPI.Linked",
            "id": self.__next_id(),
            "params": [token, lambda, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise QueuesAPIError.from_json('linked', payload['error'])
        return [Queue.from_json(x) for x in (payload['result'] or [])]

    async def list(self, token: Any) -> List[Queue]:
        """
        List of all queues
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "QueuesAPI.List",
            "id": self.__next_id(),
            "params": [token, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise QueuesAPIError.from_json('list', payload['error'])
        return [Queue.from_json(x) for x in (payload['result'] or [])]

    async def assign(self, token: Any, name: str, lambda: str) -> bool:
        """
        Assign lambda to queue (re-link)
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "QueuesAPI.Assign",
            "id": self.__next_id(),
            "params": [token, name, lambda, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise QueuesAPIError.from_json('assign', payload['error'])
        return payload['result']

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class QueuesAPIBatch:
    """
    API for managing queues
    """

    def __init__(self, client: QueuesAPIClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def create(self, token: Any, queue: Queue):
        """
        Create queue and link it to lambda and start worker
        """
        params = [token, queue.to_json(), ]
        method = "QueuesAPI.Create"
        self.__add_request(method, params, lambda payload: Queue.from_json(payload))

    def remove(self, token: Any, name: str):
        """
        Remove queue and stop worker
        """
        params = [token, name, ]
        method = "QueuesAPI.Remove"
        self.__add_request(method, params, lambda payload: payload)

    def linked(self, token: Any, lambda: str):
        """
        Linked queues for lambda
        """
        params = [token, lambda, ]
        method = "QueuesAPI.Linked"
        self.__add_request(method, params, lambda payload: [Queue.from_json(x) for x in (payload or [])])

    def list(self, token: Any):
        """
        List of all queues
        """
        params = [token, ]
        method = "QueuesAPI.List"
        self.__add_request(method, params, lambda payload: [Queue.from_json(x) for x in (payload or [])])

    def assign(self, token: Any, name: str, lambda: str):
        """
        Assign lambda to queue (re-link)
        """
        params = [token, name, lambda, ]
        method = "QueuesAPI.Assign"
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
                raise QueuesAPIError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans
from aiohttp import client


from typing import Any



class UserAPIError(RuntimeError):
    def __init__(self, method: str, code: int, message: str, data: Any):
        super().__init__('{}: {}: {} - {}'.format(method, code, message, data))
        self.code = code
        self.message = message
        self.data = data

    @staticmethod
    def from_json(method: str, payload: dict) -> 'UserAPIError':
        return UserAPIError(
            method=method,
            code=payload['code'],
            message=payload['message'],
            data=payload.get('data')
        )


class UserAPIClient:
    """
    User/admin profile API
    """

    def __init__(self, base_url: str = 'https://127.0.0.1:3434/u/', session: Optional[client.ClientSession] = None):
        self.__url = base_url
        self.__id = 1
        self.__request = session.request if session is not None else client.request

    def __next_id(self):
        self.__id += 1
        return self.__id

    async def login(self, login: str, password: str) -> Any:
        """
        Login user by username and password. Returns signed JWT
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "UserAPI.Login",
            "id": self.__next_id(),
            "params": [login, password, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise UserAPIError.from_json('login', payload['error'])
        return payload['result']

    async def change_password(self, token: Any, password: str) -> bool:
        """
        Change password for the user
        """
        response = await self._invoke({
            "jsonrpc": "2.0",
            "method": "UserAPI.ChangePassword",
            "id": self.__next_id(),
            "params": [token, password, ]
        })
        assert response.status // 100 == 2, str(response.status) + " " + str(response.reason)
        payload = await response.json()
        if 'error' in payload:
            raise UserAPIError.from_json('change_password', payload['error'])
        return payload['result']

    async def _invoke(self, request):
        return await self.__request('POST', self.__url, json=request)


class UserAPIBatch:
    """
    User/admin profile API
    """

    def __init__(self, client: UserAPIClient, size: int = 10):
        self.__id = 1
        self.__client = client
        self.__requests = []
        self.__batch = {}
        self.__batch_size = size

    def __next_id(self):
        self.__id += 1
        return self.__id

    def login(self, login: str, password: str):
        """
        Login user by username and password. Returns signed JWT
        """
        params = [login, password, ]
        method = "UserAPI.Login"
        self.__add_request(method, params, lambda payload: payload)

    def change_password(self, token: Any, password: str):
        """
        Change password for the user
        """
        params = [token, password, ]
        method = "UserAPI.ChangePassword"
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
                raise UserAPIError.from_json(request['method'], payload['error'])
            else:
                ans.append(factory(payload['result']))
        return ans
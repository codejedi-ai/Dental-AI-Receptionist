import logging
from typing import List, Optional

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

from app.core.tools import Tool, ToolRegistry, ToolSpec

logger = logging.getLogger("mcp-registry")


class MCPTool(Tool):
    def __init__(self, session: ClientSession, name: str, description: str, parameters: dict):
        self._session = session
        self._spec = ToolSpec(name=name, description=description, parameters=parameters)

    @property
    def spec(self) -> ToolSpec:
        return self._spec

    async def __call__(self, **kwargs) -> str:
        result = await self._session.call_tool(self._spec.name, kwargs)
        output = []
        for content in result.content:
            if hasattr(content, "text"):
                output.append(content.text)
            elif hasattr(content, "image"):
                output.append("[Image Content]")
            else:
                output.append(str(content))
        return "\n".join(output)


class MCPToolRegistry(ToolRegistry):
    def __init__(self, command: str, args: List[str]):
        self.command = command
        self.args = args
        self.session: Optional[ClientSession] = None
        self._client_ctx = None
        self._tools: List[Tool] = []

    async def start(self) -> None:
        logger.info("Connecting to MCP server: %s %s", self.command, " ".join(self.args))
        server_params = StdioServerParameters(command=self.command, args=self.args)
        self._client_ctx = stdio_client(server_params)

        read, write = await self._client_ctx.__aenter__()
        self.session = ClientSession(read, write)
        await self.session.__aenter__()
        await self.session.initialize()

        tools_resp = await self.session.list_tools()
        self._tools = []
        for tool in tools_resp.tools or []:
            self._tools.append(
                MCPTool(
                    session=self.session,
                    name=tool.name,
                    description=tool.description or "",
                    parameters=getattr(tool, "inputSchema", {}) or {},
                )
            )

    def list_tools(self) -> List[Tool]:
        return list(self._tools)

    async def stop(self) -> None:
        if self.session:
            await self.session.__aexit__(None, None, None)
        if self._client_ctx:
            await self._client_ctx.__aexit__(None, None, None)
        self.session = None
        self._client_ctx = None
        self._tools = []

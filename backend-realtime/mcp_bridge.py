import logging
import json
import asyncio
from typing import Any, Dict, List, Optional
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from livekit.agents import llm

logger = logging.getLogger("mcp-bridge")

class MCPBridge:
    def __init__(self, command: str, args: List[str]):
        self.command = command
        self.args = args
        self.session: Optional[ClientSession] = None
        self._client_ctx = None
        self.fnc_ctx = llm.FunctionContext()

    async def start(self):
        """Connect to MCP server and discover tools."""
        logger.info(f"Connecting to MCP server: {self.command} {' '.join(self.args)}")
        server_params = StdioServerParameters(command=self.command, args=self.args)
        self._client_ctx = stdio_client(server_params)
        
        # Connect to the server
        read, write = await self._client_ctx.__aenter__()
        self.session = ClientSession(read, write)
        await self.session.__aenter__()
        await self.session.initialize()
        
        logger.info("Connected to MCP server. Discovering tools...")
        tools_resp = await self.session.list_tools()
        
        if not tools_resp.tools:
            logger.warning("No tools discovered on MCP server.")
        else:
            for tool in tools_resp.tools:
                self._register_tool(tool)
                logger.info(f"Registered MCP tool: {tool.name}")
        
        logger.info("MCP Bridge initialization complete.")

    def _register_tool(self, tool):
        """Dynamically register an MCP tool as a LiveKit function."""
        
        # Define a wrapper that matches the MCP tool's expected signature
        # Since we're dynamic, we use a generic wrapper but provide the schema
        async def mcp_tool_wrapper(**kwargs):
            if not self.session:
                return "Error: MCP session not initialized."
            
            logger.info(f"Calling MCP tool {tool.name} with arguments: {kwargs}")
            try:
                result = await self.session.call_tool(tool.name, kwargs)
                
                output = []
                for content in result.content:
                    if hasattr(content, 'text'):
                        output.append(content.text)
                    elif hasattr(content, 'image'):
                        output.append("[Image Content]")
                    else:
                        output.append(str(content))
                
                return "\n".join(output)
            except Exception as e:
                logger.error(f"Error calling MCP tool {tool.name}: {e}")
                return f"Error: {str(e)}"

        # LiveKit's register_fnc can take an override schema if needed,
        # but the easiest way for dynamic tools is to use the low-level API
        # or rely on the LLM knowing the tool exists if we can't easily sync schemas.
        # For now, let's use a simpler approach: 
        # We'll use the docstring to help the LLM, and since we use **kwargs,
        # some LLMs are smart enough to guess parameters from descriptions.
        
        self.fnc_ctx.register_fnc(
            mcp_tool_wrapper,
            name=tool.name,
            description=tool.description
        )

    async def stop(self, *args, **kwargs):
        """Cleanup connection."""
        if self.session:
            await self.session.__aexit__(None, None, None)
        if self._client_ctx:
            await self._client_ctx.__aexit__(None, None, None)
        logger.info("MCP bridge stopped.")

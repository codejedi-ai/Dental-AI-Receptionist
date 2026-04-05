from typing import AsyncGenerator, List, Optional
from app.core.interfaces import Brain
from app.core.llm_client import LLMClient, Message
from app.core.tools import ToolRegistry

class LLMBrain(Brain):
    """
    A brain powered by an LLM.
    This class is runtime-agnostic; it relies on an injected LLMClient.
    """
    def __init__(
        self,
        system_prompt: str,
        client: LLMClient,
        tool_registry: Optional[ToolRegistry] = None,
    ):
        self.client = client
        self.system_prompt = system_prompt
        self.tool_registry = tool_registry
        self.messages: List[Message] = []
        if system_prompt:
            self.messages.append(Message(role="system", content=system_prompt))

    async def think(self, user_input: str) -> AsyncGenerator[str, None]:
        self.messages.append(Message(role="user", content=user_input))

        full_response = ""
        async for content in self.client.stream_chat(
            self.messages,
            tool_registry=self.tool_registry,
        ):
            if content:
                full_response += content
                yield content

        self.messages.append(Message(role="assistant", content=full_response))

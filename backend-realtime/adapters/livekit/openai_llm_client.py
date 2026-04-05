import os
from typing import AsyncGenerator, Iterable, Optional

from livekit.plugins import openai
from livekit.agents import llm

from app.core.llm_client import LLMClient, Message
from app.core.tools import ToolRegistry
from app.adapters.livekit.tool_registry_adapter import build_function_context


class LiveKitOpenAIClient(LLMClient):
    """
    LiveKit-backed OpenAI client adapter.
    """
    def __init__(self, model: Optional[str] = None, base_url: Optional[str] = None):
        self.llm_plugin = openai.LLM(
            model=model or os.getenv("LLM_MODEL", "gpt-4o-mini"),
            base_url=base_url or os.getenv("LLM_BASE_URL", None),
        )

    async def stream_chat(
        self,
        messages: Iterable[Message],
        tool_registry: Optional[ToolRegistry] = None,
    ) -> AsyncGenerator[str, None]:
        chat_context = llm.ChatContext()
        for message in messages:
            chat_context.append(role=message.role, text=message.content)

        fnc_ctx = None
        if tool_registry is not None:
            fnc_ctx = build_function_context(tool_registry)

        stream = await self.llm_plugin.chat(history=chat_context, fnc_ctx=fnc_ctx)
        async for chunk in stream:
            content = chunk.choices[0].delta.content
            if content:
                yield content

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import AsyncGenerator, Iterable, Optional

from app.core.tools import ToolRegistry


@dataclass(frozen=True)
class Message:
    role: str
    content: str


class LLMClient(ABC):
    """
    Transport-agnostic LLM client interface.
    Keeps the Brain decoupled from any specific runtime (LiveKit, CLI, etc.).
    """
    @abstractmethod
    async def stream_chat(
        self,
        messages: Iterable[Message],
        tool_registry: Optional[ToolRegistry] = None,
    ) -> AsyncGenerator[str, None]:
        pass

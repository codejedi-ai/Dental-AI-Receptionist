from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Dict, List


@dataclass(frozen=True)
class ToolSpec:
    name: str
    description: str
    parameters: Dict[str, Any]


class Tool(ABC):
    @property
    @abstractmethod
    def spec(self) -> ToolSpec:
        pass

    @abstractmethod
    async def __call__(self, **kwargs) -> str:
        pass


class ToolRegistry(ABC):
    @abstractmethod
    def list_tools(self) -> List[Tool]:
        pass

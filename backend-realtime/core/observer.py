from dataclasses import dataclass
from typing import Any, Dict, List, Protocol


@dataclass(frozen=True)
class AgentEvent:
    type: str
    payload: Dict[str, Any]


class Observer(Protocol):
    def on_event(self, event: AgentEvent) -> None:
        pass


class Observable:
    def __init__(self) -> None:
        self._observers: List[Observer] = []

    def add_observer(self, observer: Observer) -> None:
        if observer not in self._observers:
            self._observers.append(observer)

    def remove_observer(self, observer: Observer) -> None:
        if observer in self._observers:
            self._observers.remove(observer)

    def notify(self, event: AgentEvent) -> None:
        for observer in list(self._observers):
            observer.on_event(event)

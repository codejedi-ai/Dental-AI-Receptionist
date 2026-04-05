from abc import ABC, abstractmethod

from app.core.observer import Observable


class AgentController(Observable, ABC):
    """
    Controller layer: orchestrates Model (Brain) and View (Ear/Voice).
    """
    def __init__(self) -> None:
        super().__init__()

    @abstractmethod
    async def run(self, *args, **kwargs):
        pass

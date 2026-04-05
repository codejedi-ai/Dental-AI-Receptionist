from abc import ABC, abstractmethod
from typing import AsyncGenerator

class Brain(ABC):
    """
    The Model: Pure logic.
    Takes a string input (user message) and returns a string output (agent response).
    """
    @abstractmethod
    async def think(self, user_input: str) -> AsyncGenerator[str, None]:
        """
        Process the user input and yield chunks of the response.
        Using a generator allows for streaming responses (low latency).
        """
        pass

class Voice(ABC):
    """
    The View (Audio): Handles Text-to-Speech configuration.
    """
    @abstractmethod
    def get_tts_plugin(self):
        """Returns the configured TTS plugin instance."""
        pass

class Ear(ABC):
    """
    The View (Input): Handles Speech-to-Text configuration.
    """
    @abstractmethod
    def get_stt_plugin(self):
        """Returns the configured STT plugin instance."""
        pass

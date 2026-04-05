import json
import os
from dataclasses import dataclass, field
from typing import Awaitable, Callable, List, Optional

from app.factory.component_factory import ComponentFactory


AsyncHook = Callable[[], Awaitable[None]]


@dataclass
class AgentAssembly:
    ear: object
    brain: object
    voice: object
    intro_phrase: str
    startup_hooks: List[AsyncHook] = field(default_factory=list)
    shutdown_hooks: List[AsyncHook] = field(default_factory=list)

class AgentFactory:
    @staticmethod
    def load_config(config_name: str) -> dict:
        """
        Loads the JSON configuration for the specified agent.
        Accepts a name (resolved from app/configs) or a direct .json path.
        """
        base_dir = os.path.dirname(os.path.abspath(__file__))
        candidate = config_name
        if not candidate.endswith(".json"):
            candidate = f"{config_name}.json"

        if os.path.isabs(candidate) and os.path.exists(candidate):
            config_path = candidate
        elif os.path.exists(candidate):
            config_path = os.path.abspath(candidate)
        else:
            config_path = os.path.join(base_dir, "configs", candidate)
        
        if not os.path.exists(config_path):
            raise FileNotFoundError(f"Configuration file not found: {config_path}")
            
        with open(config_path, "r", encoding="utf-8") as f:
            return json.load(f)

    @staticmethod
    def create_agent(config_name: str = "celeste") -> AgentAssembly:
        """
        Factory method to assemble an agent from a JSON configuration.
        """
        config = AgentFactory.load_config(config_name)
        
        ear = ComponentFactory.create_ear(config.get("ear", {}))
        voice = ComponentFactory.create_voice(config.get("voice", {}))
        brain = ComponentFactory.create_brain(config.get("brain", {}))

        startup_hooks: List[AsyncHook] = []
        shutdown_hooks: List[AsyncHook] = []

        tool_registry = getattr(brain, "tool_registry", None)
        if tool_registry is not None:
            if hasattr(tool_registry, "start"):
                startup_hooks.append(tool_registry.start)
            if hasattr(tool_registry, "stop"):
                shutdown_hooks.append(tool_registry.stop)

        return AgentAssembly(
            ear=ear,
            brain=brain,
            voice=voice,
            intro_phrase=config.get("intro_phrase", "Hello."),
            startup_hooks=startup_hooks,
            shutdown_hooks=shutdown_hooks,
        )

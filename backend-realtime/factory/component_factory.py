from typing import Optional, Tuple

from app.brains.eliza_brain import ElizaBrain
from app.brains.llm_brain import LLMBrain
from app.core.tools import ToolRegistry
from app.ears.openai_ear import OpenAIEar
from app.voices.rime_voice import RimeVoice
from app.adapters.livekit.openai_llm_client import LiveKitOpenAIClient
from app.tools.mcp_registry import MCPToolRegistry


class ComponentFactory:
    @staticmethod
    def create_ear(config: dict):
        ear_type = (config or {}).get("type", "openai")
        if ear_type == "openai":
            return OpenAIEar()
        return OpenAIEar()

    @staticmethod
    def create_voice(config: dict):
        voice_type = (config or {}).get("type", "rime")
        if voice_type == "rime":
            return RimeVoice(
                speaker=config.get("speaker", "celeste"),
                model=config.get("model", "arcana"),
                speed=config.get("speed", 1.0),
            )
        return RimeVoice()

    @staticmethod
    def create_llm_client(config: dict):
        client_type = (config or {}).get("type", "livekit_openai")
        if client_type == "livekit_openai":
            return LiveKitOpenAIClient(
                model=config.get("model"),
                base_url=config.get("base_url"),
            )
        return LiveKitOpenAIClient()

    @staticmethod
    def create_tool_registry(config: dict) -> Optional[ToolRegistry]:
        mcp_cfg = (config or {}).get("mcp", {})
        if not mcp_cfg.get("enabled"):
            return None
        command = mcp_cfg.get("command")
        args = mcp_cfg.get("args", [])
        if not command:
            return None
        return MCPToolRegistry(command=command, args=args)

    @staticmethod
    def create_brain(config: dict):
        brain_type = (config or {}).get("type", "llm")
        if brain_type == "eliza":
            return ElizaBrain()

        if brain_type == "llm":
            system_prompt = config.get("system_prompt", "")
            client_cfg = config.get("client", {})
            if "model" not in client_cfg and "model" in config:
                client_cfg = {**client_cfg, "model": config.get("model")}
            tool_registry = ComponentFactory.create_tool_registry(config.get("tools", {}))
            client = ComponentFactory.create_llm_client(client_cfg)
            return LLMBrain(
                system_prompt=system_prompt,
                client=client,
                tool_registry=tool_registry,
            )

        return ElizaBrain()

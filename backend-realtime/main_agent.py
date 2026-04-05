import logging
import os
from dotenv import load_dotenv
from livekit.agents import (
    JobContext,
    JobProcess,
    WorkerOptions,
    cli,
)
from livekit.plugins import silero
from app.agent_factory import AgentFactory
from app.adapters.livekit.livekit_controller import LiveKitController

load_dotenv()

# --- Main Entrypoint ---
async def entrypoint(ctx: JobContext):
    config_name = os.getenv("AGENT_CONFIG_PATH") or os.getenv("AGENT_CONFIG", "ludia")
    assembly = AgentFactory.create_agent(config_name=config_name)

    try:
        for hook in assembly.startup_hooks:
            await hook()

        controller = LiveKitController(
            ear=assembly.ear,
            brain=assembly.brain,
            voice=assembly.voice,
            intro_phrase=assembly.intro_phrase,
        )
        await controller.run(ctx)
    finally:
        for hook in assembly.shutdown_hooks:
            await hook()

def prewarm(proc: JobProcess):
    proc.userdata["vad"] = silero.VAD.load()

if __name__ == "__main__":
    cli.run_app(
        WorkerOptions(
            entrypoint_fnc=entrypoint,
            prewarm_fnc=prewarm,
        ),
    )

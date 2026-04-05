import os
import logging

# Disable Xet / HF Transfer
os.environ["HF_HUB_ENABLE_HF_TRANSFER"] = "0"
os.environ["HUGGINGFACE_HUB_ENABLE_XET"] = "0"

from livekit.agents import (
    JobContext,
    WorkerOptions,
    cli
)
from livekit.plugins import (
    openai,
    noise_cancellation,
    rime,
    silero,
)
from livekit.plugins.turn_detector.multilingual import MultilingualModel

# Configure logging to reduce noise
logging.basicConfig(level=logging.INFO)
logging.getLogger("urllib3").setLevel(logging.WARNING)
logging.getLogger("filelock").setLevel(logging.WARNING)

logger = logging.getLogger("warmup")

async def entrypoint(ctx: JobContext):
    # This entrypoint is not used during download-files
    pass

def prewarm(proc):
    pass

if __name__ == "__main__":
    # This script is used solely to trigger the 'download-files' command
    # for the plugins imported above during the Docker build.
    # By isolating imports here, we can cache the model download layer
    # separately from the main agent code.
    cli.run_app(
        WorkerOptions(
            entrypoint_fnc=entrypoint,
            prewarm_fnc=prewarm,
        ),
    )

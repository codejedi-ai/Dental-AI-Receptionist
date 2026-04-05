import os
import sys
from dotenv import load_dotenv

# Disable Xet / HF Transfer
os.environ["HF_HUB_ENABLE_HF_TRANSFER"] = "0"
os.environ["HUGGINGFACE_HUB_ENABLE_XET"] = "0"

# Fix imports
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from main_agent import cli, entrypoint, prewarm, WorkerOptions

# Load environment variables from .env file
dotenv_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), ".env")
load_dotenv(dotenv_path=dotenv_path)

def generate_token():
    api_key = os.getenv("LIVEKIT_API_KEY")
    api_secret = os.getenv("LIVEKIT_API_SECRET")
    if api_key and api_secret:
        from livekit import api
        token = api.AccessToken(api_key, api_secret) \
            .with_identity("human-user") \
            .with_name("Human") \
            .with_grants(api.VideoGrants(
                room_join=True,
                room="test-room",
                can_publish=True,
                can_subscribe=True,
                can_publish_data=True,
            ))
        token_jwt = token.to_jwt()
        
        # Save to shared directory (root level shared folder)
        shared_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "shared")
        os.makedirs(shared_dir, exist_ok=True)
        token_path = os.path.join(shared_dir, "token.txt")
        
        with open(token_path, "w") as f:
            f.write(token_jwt)
        print(f"Token generated and saved to {token_path}")
        print(f"\n--- Connect via Playground ---")
        print(f"https://agents-playground.livekit.io/#token={token_jwt}")
        print(f"------------------------------\n")

if __name__ == "__main__":
    generate_token()
    cli.run_app(
        WorkerOptions(
            entrypoint_fnc=entrypoint,
            prewarm_fnc=prewarm,
        ),
    )

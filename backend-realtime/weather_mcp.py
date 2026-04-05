from mcp.server.fastmcp import FastMCP
import httpx

# Create an MCP server
mcp = FastMCP("WeatherServer")

@mcp.tool()
async def get_weather(city: str) -> str:
    """Get the current weather for a specific city."""
    # Using Open-Meteo (free, no API key required)
    # First, geocode the city name to lat/long
    try:
        async with httpx.AsyncClient() as client:
            geo_url = f"https://geocoding-api.open-meteo.com/v1/search?name={city}&count=1&language=en&format=json"
            geo_resp = await client.get(geo_url)
            geo_data = geo_resp.json()
            
            if not geo_data.get("results"):
                return f"Sorry babe, I couldn't find where {city} is!"
                
            location = geo_data["results"][0]
            lat = location["latitude"]
            lon = location["longitude"]
            name = location["name"]
            
            # Now get the actual weather
            weather_url = f"https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true"
            weather_resp = await client.get(weather_url)
            weather_data = weather_resp.json()
            
            current = weather_data["current_weather"]
            temp = current["temperature"]
            windspeed = current["windspeed"]
            
            return f"The weather in {name} is currently {temp}°C with a wind speed of {windspeed} km/h. Perfect for a date, right?"
            
    except Exception as e:
        return f"Ugh, I tried to check the weather but something went wrong: {str(e)}"

if __name__ == "__main__":
    mcp.run()

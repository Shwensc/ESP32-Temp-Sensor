from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from datetime import datetime, timedelta, timezone
import sqlite3
import pandas as pd
from collections import Counter

# Initialize the FastAPI app
app = FastAPI()

# Allow CORS for specific origins
origins = [
    "http://localhost",  # Adjust this to your frontend's URL
    "http://localhost:3000",  # If using a local React development server
    "http://localhost:5173"
    # Add more origins as necessary
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,  # List of allowed origins
    allow_credentials=True,
    allow_methods=["*"],  # Allows all HTTP methods
    allow_headers=["*"],  # Allows all headers
)

# SQLite Database connection
DATABASE = "../temperature.db"

# Helper function to connect to the SQLite database
def get_db_connection():
    conn = sqlite3.connect(DATABASE)
    conn.row_factory = sqlite3.Row
    return conn

# Define the possible time ranges
TIME_RANGES = {
    "1h": timedelta(hours=1),
    "12h": timedelta(hours=12),
    "1d": timedelta(days=1),
    "1w": timedelta(weeks=1),
}

# Function to get the temperatures for a given time range
def get_temperatures_in_range(range: str):
    if range not in TIME_RANGES:
        raise HTTPException(status_code=400, detail="Invalid range. Valid values are: 1h, 12h, 1d, 1w.")
    
    time_delta = TIME_RANGES[range]
    end_time = datetime.now(timezone.utc)
    start_time = end_time - time_delta
    
    conn = get_db_connection()
    query = "SELECT * FROM temperatures WHERE created_at >= ?"
    
    # Format the start_time correctly
    formatted_start_time = start_time.strftime("%Y-%m-%d %H:%M:%S")
    
    print(f"Executing SQL query: {query} with start_time = {formatted_start_time}")
    
    rows = conn.execute(query, (formatted_start_time,)).fetchall()
    conn.close()
    
    return [dict(row) for row in rows]

# Function to calculate the statistics
def calculate_statistics(temperatures):
    if not temperatures:
        return None
    
    data = pd.DataFrame(temperatures)
    highest = data["temperature"].max()
    lowest = data["temperature"].min()
    mean = data["temperature"].mean()
    median = data["temperature"].median()
    mode = Counter(data["temperature"]).most_common(1)[0][0]
    
    return {
        "highest": highest,
        "lowest": lowest,
        "mean": mean,
        "median": median,
        "mode": mode,
    }

# Route to get statistics
@app.get("/stats")
async def get_stats(range: str = "1h"):
    temperatures = get_temperatures_in_range(range)
    stats = calculate_statistics(temperatures)
    
    if stats is None:
        raise HTTPException(status_code=404, detail="No temperature data found for the given time range.")
    
    return stats

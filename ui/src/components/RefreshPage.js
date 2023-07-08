import { useState } from "react";
import { Button } from '@mui/material';
import LinearProgressWithLabel from '@mui/material/LinearProgress';
import "./RefreshPage.css";

// Render the refresh page, which features a button to send signal to backend to re-calculate data.
export default function RefreshPage() {
    // Define an isLoading flag.
    const [isLoading, setIsLoading] = useState(false);
    const [progress, setProgress] = useState(0);
    
    document.body.style.backgroundColor = "black"
    
    // Simple function to perform async refresh transmission.
    function buttonClick() {
        setIsLoading(true)
        // Create a web socket to the backend to receive progress updates.
        var wsUrl = ((window.location.protocol === "https:") ? "wss://" : "ws://") + "localhost:5000" + "/refresh"
        console.log('Creating new web socket at URL: ' + wsUrl)
        var ws = new WebSocket(wsUrl);
        // Define the socket callback for when messages are received.
        ws.onmessage = event => {
            console.log('Progress received: ' + event.data)
            setProgress(Number(event.data))
        }
        // Define callback for when socket is closed.
        ws.onclose = event => {
            console.log('Web socket closed!')
            // setProgress(100)
            setIsLoading(false)
        }
    }

    // Return this JSX content to be rendered.
    return (
        <>
            <div className="refresh-container">
                {/* Display a button to refresh, and progress bar when loading. */}
                <Button 
                    className="refresh-button" 
                    variant="contained" 
                    color="info"
                    sx={{
                        "&.Mui-disabled": {
                          background: "grey",
                          color: "light-grey"
                        }
                      }}
                    disabled={isLoading} 
                    onClick={buttonClick}
                >
                    {(isLoading) ? 'Loading...' : 'Refresh Portfolio Data'}
                </Button>
                <LinearProgressWithLabel className="refresh-progressbar" variant="determinate" value={progress} />
            </div>
        </>
    );
}
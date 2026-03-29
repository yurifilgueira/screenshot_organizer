import {useEffect, useState} from "react";

import './App.css';

import * as runtime from "../wailsjs/runtime/runtime";

interface Screenshot {
    filename: string;
    category: string;
}

function App() {
    const [_history, setHistory] = useState<Screenshot[]>([]);
    const [status, setStatus] = useState<string>("Monitoring folder...");

    useEffect(() => {
        runtime.EventsOn("processing-start", (path: string) => {
            setStatus(`Analyzing: ${path.split('\\').pop()}`);
        });

        runtime.EventsOn("new-result", (result: Screenshot) => {
            setHistory((prev) => [result, ...prev]);
            setStatus("Monitoring folder...");
        });

        return () => {
            runtime.EventsOff("processing-start");
            runtime.EventsOff("new-result");
        }
    }, []);

    return (
        <div className="App">
            <header>
                <h1>Screenshot Organizer</h1>
                <p className={'status-badge'}>{status}</p>
            </header>

            <main>
                <section className="config-section">
                    <form className="config-form" onSubmit={(e) => e.preventDefault()}>
                        <div className="form-group">
                            <label htmlFor="dir-path">Screenshot Directory Path</label>
                            <input 
                                id="dir-path"
                                type="text" 
                                placeholder="C:\Users\Name\Pictures\Screenshots"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="api-key">Gemini API Key</label>
                            <input 
                                id="api-key"
                                type="password" 
                                placeholder="Enter your Gemini API Key"
                            />
                        </div>
                        <button type="submit" className="save-button">Save Configuration</button>
                    </form>
                </section>

                <div className="history-list">
                    {_history.map((item, index) => (
                        <div key={index} className="card">
                            <span className="filename">{item.filename}</span>
                            <span className="category-tag">{item.category}</span>
                        </div>
                    ))}
                </div>
            </main>
        </div>
    )
}

export default App;
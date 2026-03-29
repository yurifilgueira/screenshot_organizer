import {useEffect, useState} from "react";

import './App.css';

import * as runtime from "../wailsjs/runtime/runtime";
import {LogError, LogInfo} from "../wailsjs/runtime";
import {SaveConfig, SelectDirectory} from "../wailsjs/go/main/App";

interface Screenshot {
    filename: string;
    category: string;
}

function App() {
    const [_history, setHistory] = useState<Screenshot[]>([]);
    const [status, setStatus] = useState<string>("Monitoring folder...");
    const [dirPath, setDirPath] = useState<string>("");
    const [apiKey, setApiKey] = useState<string>("");

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

    const handleSelectFolder = async () => {
        const selected = await SelectDirectory();
        if (selected) {
            setDirPath(selected);
        }
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        LogInfo(`Saving config: configData`);

        if (!dirPath || dirPath.length === 0 || !apiKey || apiKey.length === 0) {
            return;
        }

        try {
            SaveConfig(dirPath, apiKey);
        } catch (err) {
            LogError(`Error saving config: ${err}`);
        }
    };

    return (
        <div className="App">
            <header>
                <h1>Screenshot Organizer</h1>
                <p className={'status-badge'}>{status}</p>
            </header>

            <main>
                <section className="config-section">
                    <form className="config-form" onSubmit={handleSubmit}>
                        <div className="form-group">
                            <label htmlFor="dir-path">Screenshot Directory Path</label>
                            <div className="input-with-button">
                                <input 
                                    id="dir-path"
                                    type="text" 
                                    value={dirPath}
                                    onChange={(e) => setDirPath(e.target.value)}
                                    placeholder="C:\Users\Name\Pictures\Screenshots"
                                />
                                <button 
                                    type="button" 
                                    className="browse-button"
                                    onClick={handleSelectFolder}
                                >
                                    Browse
                                </button>
                            </div>
                        </div>
                        <div className="form-group">
                            <label htmlFor="api-key">Gemini API Key</label>
                            <input 
                                id="api-key"
                                type="password" 
                                value={apiKey}
                                onChange={(e) => setApiKey(e.target.value)}
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
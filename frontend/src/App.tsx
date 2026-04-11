import {useEffect, useState} from "react";

import './App.css';

import * as runtime from "../wailsjs/runtime/runtime";
import {LogError, LogInfo} from "../wailsjs/runtime";
import {GetConfig, SaveConfig, SelectDirectory, SetHideOnClose} from "../wailsjs/go/main/App";

interface Screenshot {
    filename: string;
    category: string;
}

const EyeIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
        <circle cx="12" cy="12" r="3"></circle>
    </svg>
);

const EyeOffIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
        <line x1="1" y1="1" x2="23" y2="23"></line>
    </svg>
);

function App() {
    const [_history, setHistory] = useState<Screenshot[]>([]);
    const [status, setStatus] = useState<string>("Monitoring folder...");
    const [dirPath, setDirPath] = useState<string>("");
    const [apiKey, setApiKey] = useState<string>("");
    const [showPassword, setShowPassword] = useState<boolean>(false);
    const [hideOnClose, setHideOnClose] = useState<boolean>(false);

    useEffect(() => {
        const fetchConfig = async () => {
            const config = await GetConfig();
            if (config) {
                setDirPath(config.screenshotsDirPath);
                setApiKey(config.apiKey);
            }
        };

        fetchConfig();

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

    const handleToggle = (e: { target: any; }) => {
        const target = e.target;
        setHideOnClose(target.checked);

        SetHideOnClose(target.checked);

        LogInfo(`Setting hideOnClose to ${target.checked}`);
    }

    return (
        <div className="App">
            <header>
                <h1>Screenshot Organizer</h1>
                <p className={'status-badge'}>{status}</p>
            </header>

            <main>
                <section className="section">
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
                                    autoComplete="off"
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
                            <div className="input-with-button">
                                <input 
                                    id="api-key"
                                    type={showPassword ? "text" : "password"} 
                                    value={apiKey}
                                    onChange={(e) => setApiKey(e.target.value)}
                                    placeholder="Enter your Gemini API Key"
                                />
                                <button 
                                    type="button" 
                                    className="toggle-password-button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    title={showPassword ? "Hide" : "Show"}
                                >
                                    {showPassword ? <EyeOffIcon /> : <EyeIcon />}
                                </button>
                            </div>
                        </div>
                        <button type="submit" className="save-button">Save Configuration</button>
                    </form>
                </section>
                <section className="section">
                    <div className="form-group">
                        <label>Minimize on Close Window</label>
                        <label className="toggle">
                            <input type="checkbox" id="btnToggle" name="btnToggle" checked={hideOnClose} onChange={handleToggle}/>
                            <span className="slider"></span>
                        </label>
                    </div>
                </section>
            </main>
        </div>
    )
}

export default App;
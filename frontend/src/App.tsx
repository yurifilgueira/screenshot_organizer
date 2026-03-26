import {useEffect, useState} from "react";

import * as runtime from "../wailsjs/runtime/runtime";

interface Screenshot {
    filename: string;
    category: string;
}

function App() {
    const [history, setHistory] = useState<Screenshot[]>([]);
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
                <div className={'history-list'}>
                    {history.map((item, index) => (
                        <div key={index} className={'card'}>
                            <span className={'fileName'}>{item.filename}</span>
                            <span className={'category-tag'}>{item.category}</span>
                        </div>
                    ))}
                </div>
            </main>
        </div>
    )
}

export default App;
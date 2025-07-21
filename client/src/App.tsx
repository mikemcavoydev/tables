import { useEffect } from "react";
import "./App.css";
import { Env } from "./Env";

function App() {
    useEffect(() => {
        fetch(`${Env.API_BASE_URL}/tables`)
            .then((res) => res.json())
            .then((data) => console.log(data));
    }, []);

    return (
        <div>
            <h1>Hello World!</h1>
        </div>
    );
}

export default App;

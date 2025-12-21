import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { TDSMobileBedrockProvider } from "@toss-design-system/mobile-bedrock";

import App from "./app/App";
import "./styles/global.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <TDSMobileBedrockProvider>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </TDSMobileBedrockProvider>
  </React.StrictMode>
);

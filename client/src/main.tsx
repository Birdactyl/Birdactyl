import React, { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import AuthPage from './pages/AuthPage'
import ConsolePage from './pages/ConsolePage'
import { ProtectedRoute, NotificationContainer, useNotifications } from './components'
import { initAuth } from './lib/auth'
import { initPluginHost } from './lib/pluginHost'
import { loadAllPlugins } from './lib/pluginLoader'
import './index.css'

initPluginHost();

function App() {
  const { notifications, removeNotification } = useNotifications();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const init = async () => {
      await initAuth();
      await loadAllPlugins().catch(() => {});
      setReady(true);
    };
    init();
  }, []);

  if (!ready) {
    return <div className="w-screen h-screen bg-[#0a0a0a]" />;
  }

  return (
    <>
      <NotificationContainer notifications={notifications} onClose={removeNotification} />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/auth" element={<AuthPage />} />
        <Route path="/console/*" element={<ProtectedRoute><ConsolePage /></ProtectedRoute>} />
      </Routes>
    </>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
)

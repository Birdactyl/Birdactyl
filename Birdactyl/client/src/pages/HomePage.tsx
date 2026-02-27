import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export default function HomePage() {
  const navigate = useNavigate();

  useEffect(() => {
    navigate('/auth');
  }, [navigate]);

  return <div className="w-screen h-screen bg-[#0a0a0a]" />;
}

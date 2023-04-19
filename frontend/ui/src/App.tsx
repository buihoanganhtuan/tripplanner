import { useState } from 'react'
import { LoginPane } from './components/LoginPane'

function App() {
  const [count, setCount] = useState<number>(0)

  function onClick(e: React.MouseEvent<HTMLElement>) : void {
    setCount(curCount => curCount + 1)
  }

  return (
    <div className="App bg-gradient-to-b from-indigo-500 to-cyan-500 h-screen font-sans">
      <div className="text-7xl text-emerald-200 text-center py-11">
        Welcome to Trip Planner
      </div>
      <LoginPane />
    </div>
  )
}

export default App

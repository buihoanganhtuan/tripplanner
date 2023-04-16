import { useState } from 'react'
import { LoginPane } from './components/LoginPane'

function App() {
  const [count, setCount] = useState<number>(0)

  function onClick(e: React.MouseEvent<HTMLElement>) : void {
    setCount(curCount => curCount + 1)
  }

  return (
    <div className="App">
      <div className="shadow-lg bg-green-300">
        "Welcome to Trip Planner"
      </div>
      <LoginPane />
    </div>
  )
}

export default App

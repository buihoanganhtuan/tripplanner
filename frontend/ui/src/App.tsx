import { useState } from 'react'

function App() {
  const [count, setCount] = useState<number>(0)

  function onClick(e: React.MouseEvent<HTMLElement>) : void {
    setCount(curCount => curCount + 1)
  }

  return (
    <div className="App">
      <div className="shadow-lg bg-green-300">
        "Hello"
      </div>
      <button onClick={onClick}/>
    </div>
  )
}

export default App

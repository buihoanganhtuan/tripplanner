import { useState } from 'react'
import { AnonTripPane } from './components/AnonTripPane'
import { LoginPane } from './components/LoginPane'

export const hostname = "localhost:80"

function App() {

  return (
    <div className="App grid content-between grid-rows-landing gap-y-5 justify-items-center items-center bg-gradient-to-b from-indigo-500 to-cyan-500 bg-cover min-h-screen font-sans">
      <div className="row-start-1">
        Navbar goes here
      </div>

      <div className="row-start-2 text-7xl text-emerald-200 text-center py-11">
        Welcome to Trip Planner
      </div>

      <div className="row-start-3 grid grid-rows-2 gap-y-5 md:gap-x-10 md:grid-rows-1 md:grid-cols-1">
        <div className="row-start-1 md:col-start-1">
          <LoginPane />
        </div>
        <div className="row-start-2 md:row-start-1 md:col-start-2">
          <AnonTripPane />
        </div>
      </div>
      
    </div>
  )
}

export default App

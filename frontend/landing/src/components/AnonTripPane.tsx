import React from "react"
import { hostname } from "../App"

export function AnonTripPane() {
    function onClickCreate(e : React.MouseEvent<HTMLButtonElement>) {
        
    }

    function onSubmitGet(e: React.FormEvent<HTMLFormElement>) {
        let ri : RequestInit = {
            method: "GET",
            credentials: "omit",
            cache: "no-cache",
        }
        let tid = (e.currentTarget.elements.namedItem("tripId") as HTMLInputElement).value
        
        console.log(hostname + "/trips/" + tid)
        e.preventDefault()
        let f = async function() {
            try {
                let resp = await fetch("http://" + hostname + "/trips/" + tid, ri)
                if (resp.status < 200 || resp.status >= 300) {
                    
                }
            } catch (error) {
                console.log(error)
            }
        }
        f()

    }

    return (
        <div className="min-w-[350px] grid grid-rows-2 gap-y-2 item-center border-slate-200 rounded-lg bg-slate-100 drop-shadow-lg px-4 min-h-full">
            <form className="row-start-1 grid grid-rows-4 gap-y-1 rounded-lg" onSubmit={onSubmitGet}>
                <div className="row-start-1">Already have an anonymous trip?</div>               
                <label htmlFor="tripId" className="row-start-2">Trip ID</label>
                <input type="text" name="tripId" className="h-8 row-start-3"></input>
                <input type="submit" value="Fetch it" className="row-start-4 border-slate-200 rounded bg-slate-200 drop-shadow-md mx-4 h-10"></input>
            </form>
            <div className="row-start-2 grid grid-rows-2 items-center">
                <div className="row-start-1">Don't have an anonymous trip yet?</div>
                <button className="row-start-2 border-slate-200 rounded bg-slate-200 drop-shadow-md mx-4 h-10" onClick={onClickCreate}>Create one</button>
            </div>
        </div>
    )
}
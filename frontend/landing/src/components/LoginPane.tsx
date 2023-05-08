import React, { useState } from "react"

export function LoginPane() {
    // const [idState, setIdState] = useState<string>("")
    const [valid, setValid] = useState<string>('valid')

    const validityColors = new Map<string, string>([
        ['invalid', 'bg-pink-300'],
        ['valid', 'bg-white']
    ])

    function onChange(e: React.KeyboardEvent<HTMLInputElement>) {
        let s = e.key
        if (s.length != 1 || s[0] >= 'a' && s[0] <= 'z' || s[0] >= 'A' && s[0] <= 'Z' || s[0] >= '0' && s[0] <= '9' || s[0] == '@') {
            if (valid === 'invalid')
                setValid('valid')            
            return
        }
        
        // prevent the event (keystroke) from filling in the input field
        e.preventDefault()
        for (let i = 0; i < s.length; ++i) {
            setValid('invalid')
        }
    }

    function onClickLogin(e: React.ChangeEvent<HTMLButtonElement>) {
        
    }

    return (
        <div className="grid grid-rows-login-pane items-center border-slate-200 rounded-lg bg-slate-100 drop-shadow-lg px-4 min-w-[350px] min-h-full">
            <div className="row-start-1 self-justify-self-center">
                Icon goes here
            </div>
            <div className="row-start-2 grid grid-rows-login-input gap-y-1">
                    <div className="row-start-1 pe-3">
                        <label htmlFor="id">Username or Email</label>
                        { valid === "invalid" && 
                        <div className="text-xs">Can only contain [0-9a-zA-Z@]</div> 
                        }
                    </div>
                    <input type="text" required name="id" onKeyDown={onChange} className={"row-start-2 h-8 " + validityColors.get(valid)}/>
                    <label htmlFor="row-start-3 pe-3 password" className="">Password</label>
                    <input type="password" required name="password"  className="row-start-4 h-8"></input>
                    <div className="row-start-5 justify-self-end">Forgot password</div>
            </div>
            <div className="row-start-3 grid grid-cols-2 justify-evenly self-start">
                <button className="col-start-1 border-2 border-green-200 bg-green-200 drop-shadow-lg rounded-md mx-4">Login</button>
                <button className="col-start-2 border-2 border-gray-200 bg-gray-200 drop-shadow-lg rounded-md mx-4">Register</button>
            </div>
        </div>
    )
}
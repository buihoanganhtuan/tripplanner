import { useState } from "react"

export function LoginPane() {
    const [idState, setIdState] = useState<string>("")
    const [passState, setPassState] = useState<string>("")
    const [errorCount, setErrorCount] = useState<number>(0)

    function onChange(e: React.ChangeEvent<HTMLInputElement>) {
        if (!validateId(e.target.value)) {
            console.log("invalid")
            setErrorCount(cur => cur+1)
            setIdState("bg-green-300")
        }
    }

    return (
        <div className="grid grid-rows-login-pane items-center">
            <div className="row-start-1"></div>
            <div className="row-start-2 grid grid-rows-login-input gap-y-1">
                    <label htmlFor="id" className="row-start-1 pe-3 w-full">Username or Email</label>
                    <input type="text" name="id" onChange={onChange} className="row-start-2 w-80"/>
                    <label htmlFor="password" className="row-start-3 pe-3 w-full">Password</label>
                    <input type="text" name="password" className="row-start-4 bg-green-300"></input>
                    <div className="row-start-5 justify-self-end">Forgot password</div>
            </div>
            <div className="row-start-3 grid grid-cols-2 justify-evenly">
                <button className="col-start-1 border-2 rounded">Login</button>
                <button className="col-start-2 border-2 rounded">Register</button>
            </div>
        </div>
    )
}

function validateId(id: string) : boolean {
    for (let i = 0; i < id.length; ++i) {
        let c = id[i]
        if (!isDigit(c) && !isAlpha(c) && !isAllowedSpecial(c))
            return false
    }
    return true
}

function isDigit(c: string) : boolean {
    return c >= '0' && c <= '9'
}

function isAlpha(c: string) : boolean {
    return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

function isAllowedSpecial(c: string) : boolean {
    return c == '_' || c == '.'
}
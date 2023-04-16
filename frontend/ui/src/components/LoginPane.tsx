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
        <div className="">
            <div className="bg-green-300">Login</div>
            <label htmlFor="id">Username or Email</label>
            <input type="text" name="id" onChange={onChange} className={`${idState}`}/>
            <div>{"Error count " + errorCount}</div>
            <label htmlFor="password">Password</label>
            <input type="text" name="password" className="bg-green-300"></input>
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
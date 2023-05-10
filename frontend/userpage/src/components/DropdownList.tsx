import { useState } from "react"

interface DropdownBoxProps<T extends JSX.Element> {
    name: string
    children: T[]
}

interface DropdownBoxLocalState {
    collapsed: boolean
}

export function DropdownBox<T extends JSX.Element>(props: DropdownBoxProps<T>) {
    const [state, setState] = useState<DropdownBoxLocalState>({
        collapsed: false
    })

    const handleBarClick = () => {
        setState(prev => { return { ...prev, collapsed: !prev.collapsed } })
    }

    // https://stackoverflow.com/questions/47977751/css-set-parents-height-to-0-but-child-div-still-show
    return (
        <div className="grid grid-rows-drop-down justify-items-center items-center gap-y-2">
            <div onMouseDown={handleBarClick} className="row-start-1 border-2 rounded-md w-[20rem] h-[30px] bg-green-200 drop-shadow-lg grid justify-items-center">{`${props.name} (${props.children.length})`}</div>
            <div className={"row-start-2 transition-all duration-500 ease overflow-hidden " + (state.collapsed ? "max-h-40" : "max-h-0")}>
                {props.children}
            </div>
        </div>
    )
}
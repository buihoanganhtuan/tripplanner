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

    return (
        <div className="grid grid-rows-drop-down justify-items-center items-center gap-y-2">
            <div onMouseDown={handleBarClick} className="row-start-1 border-2 rounded-sm drop-shadow-lg">{`${props.name} (${props.children.length})`}</div>
            <div className="row-start-2">
                {!state.collapsed ? props.children : null}
            </div>
        </div>
    )
}
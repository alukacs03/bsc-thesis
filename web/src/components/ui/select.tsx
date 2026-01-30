import * as React from "react"
import { ChevronDown, Check } from "lucide-react"

interface SelectContextValue {
  value: string
  onValueChange: (value: string) => void
  open: boolean
  setOpen: (open: boolean) => void
  selectedLabel: string | null
  setSelectedLabel: (label: string | null) => void
}

const SelectContext = React.createContext<SelectContextValue | null>(null)

function useSelectContext() {
  const context = React.useContext(SelectContext)
  if (!context) {
    throw new Error("Select components must be used within a Select")
  }
  return context
}

interface SelectProps {
  value: string
  onValueChange: (value: string) => void
  children: React.ReactNode
}

export function Select({ value, onValueChange, children }: SelectProps) {
  const [open, setOpen] = React.useState(false)
  const [selectedLabel, setSelectedLabel] = React.useState<string | null>(null)

  React.useEffect(() => {
    setSelectedLabel(null)
  }, [value])

  return (
    <SelectContext.Provider value={{ value, onValueChange, open, setOpen, selectedLabel, setSelectedLabel }}>
      <div className="relative">
        {children}
      </div>
    </SelectContext.Provider>
  )
}

interface SelectTriggerProps {
  className?: string
  children: React.ReactNode
}

export function SelectTrigger({ className = "", children }: SelectTriggerProps) {
  const { open, setOpen } = useSelectContext()

  return (
    <button
      type="button"
      onClick={() => setOpen(!open)}
      className={`flex h-10 w-full items-center justify-between rounded-md border border-slate-200 bg-white px-3 py-2 text-sm ring-offset-white placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-950 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ${className}`}
    >
      {children}
      <ChevronDown className={`h-4 w-4 opacity-50 transition-transform ${open ? "rotate-180" : ""}`} />
    </button>
  )
}

interface SelectValueProps {
  placeholder?: string
}

export function SelectValue({ placeholder }: SelectValueProps) {
  const { value, selectedLabel } = useSelectContext()
  const displayText = selectedLabel ?? value

  return (
    <span className={`${displayText ? "text-slate-900" : "text-slate-500"} min-w-0 flex-1 truncate`}>
      {displayText || placeholder}
    </span>
  )
}

interface SelectContentProps {
  children: React.ReactNode
}

export function SelectContent({ children }: SelectContentProps) {
  const { open, setOpen } = useSelectContext()
  const ref = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        // Check if the click was on the trigger button
        const trigger = ref.current.parentElement?.querySelector('button')
        if (trigger && !trigger.contains(event.target as Node)) {
          setOpen(false)
        }
      }
    }

    if (open) {
      document.addEventListener("mousedown", handleClickOutside)
      return () => document.removeEventListener("mousedown", handleClickOutside)
    }
  }, [open, setOpen])

  return (
    <div
      ref={ref}
      aria-hidden={!open}
      className={`absolute top-full left-0 z-50 mt-1 max-h-60 min-w-[8rem] w-full overflow-auto rounded-md border border-slate-200 bg-white p-1 shadow-md ${open ? "animate-in fade-in-0 zoom-in-95" : "hidden"}`}
    >
      {children}
    </div>
  )
}

interface SelectItemProps {
  value: string
  children: React.ReactNode
}

export function SelectItem({ value: itemValue, children }: SelectItemProps) {
  const { value, onValueChange, setOpen, setSelectedLabel } = useSelectContext()
  const isSelected = value === itemValue
  const labelText = typeof children === "string" ? children : null

  React.useEffect(() => {
    if (isSelected && labelText) {
      setSelectedLabel(labelText)
    }
  }, [isSelected, labelText, setSelectedLabel])

  return (
    <div
      onClick={() => {
        onValueChange(itemValue)
        if (labelText) {
          setSelectedLabel(labelText)
        }
        setOpen(false)
      }}
      className={`relative flex w-full cursor-pointer select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none hover:bg-slate-100 focus:bg-slate-100 ${
        isSelected ? "bg-slate-100" : ""
      }`}
    >
      <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
        {isSelected && <Check className="h-4 w-4" />}
      </span>
      {children}
    </div>
  )
}

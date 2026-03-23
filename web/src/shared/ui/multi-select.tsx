import * as React from "react"
import { Check, ChevronsUpDown } from "lucide-react"

import { cn } from "@/shared/lib/utils"
import { Button } from "@/shared/ui/button"
import { Badge } from "@/shared/ui/badge"

import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from "@/shared/ui/popover"

import {
  Command,
  CommandInput,
  CommandList,
  CommandItem,
  CommandEmpty,
} from "@/shared/ui/command"

export type MultiSelectItem = {
  label: string
  value: string
}

export type MultiSelectProps = {
  items: MultiSelectItem[]
  value: string[]
  onChange: (value: string[]) => void
  placeholder?: string
  maxVisible?: number
  className?: string
}

export const MultiSelect: React.FC<MultiSelectProps> = ({
  items,
  value,
  onChange,
  placeholder = "Select...",
  maxVisible = 1,
  className,
}) => {
  const toggle = (val: string) => {
    if (value.includes(val)) {
      onChange(value.filter((v) => v !== val))
    } else {
      onChange([...value, val])
    }
  }

  const selectedItems = items.filter((i) => value.includes(i.value))

  const visible = selectedItems.slice(0, maxVisible)
  const restCount = selectedItems.length - visible.length

  return (
    <Popover>
      <PopoverTrigger>
        <Button
          variant="outline"
          className={cn(
            "w-40 justify-between rounded-[8px] px-3",
            className
          )}
        >
          <div className="flex items-center gap-1 overflow-hidden">
            {selectedItems.length === 0 && (
              <span className="text-muted-foreground text-body-small">
                {placeholder}
              </span>
            )}

            {visible.map((item) => (
              <Badge
                key={item.value}
                variant="secondary"
                className="max-w-25 truncate text-caption rounded-[6px]"
              >
                {item.label}
              </Badge>
            ))}

            {restCount > 0 && (
              <Badge
                variant="secondary"
                className="text-caption rounded-[6px]"
              >
                +{restCount}
              </Badge>
            )}
          </div>

          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-40 p-0" align="end">
        <Command>
          <CommandInput placeholder="Search..." />

          <CommandList>
            <CommandEmpty>Ничего не найдено</CommandEmpty>

            {items.map((item) => {
              const isSelected = value.includes(item.value)

              return (
                <CommandItem
                  key={item.value}
                  onSelect={() => toggle(item.value)}
                  className="cursor-pointer"
                >
                  <Check
                    className={cn(
                      "mr-2 h-4 w-4",
                      isSelected ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {item.label}
                </CommandItem>
              )
            })}
          </CommandList>
        </Command>

        {value.length > 0 && (
          <div className="border-t p-2">
            <Button
              variant="ghost"
              size="sm"
              className="w-full"
              onClick={() => onChange([])}
            >
              Очистить
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
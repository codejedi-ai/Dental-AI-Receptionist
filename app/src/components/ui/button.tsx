import { forwardRef, ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "outline" | "ghost";
  size?: "default" | "sm" | "lg";
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(({ className, variant = "default", size = "default", ...props }, ref) => {
  const variants = { default: "bg-teal-600 text-white hover:bg-teal-700", outline: "border border-gray-300 text-gray-700 hover:bg-gray-50", ghost: "hover:bg-gray-100 text-gray-700" };
  const sizes = { default: "px-4 py-2 text-sm", sm: "px-3 py-1.5 text-xs", lg: "px-6 py-3 text-base" };
  return <button ref={ref} className={cn("inline-flex items-center justify-center rounded-lg font-medium transition-colors disabled:opacity-50 disabled:pointer-events-none", variants[variant], sizes[size], className)} {...props} />;
});
Button.displayName = "Button";
export { Button };

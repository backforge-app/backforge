import { Button } from "@/shared/ui/button";
import { useSession } from "@/entities/session/model/use-session";
import { Code } from "lucide-react";

const DEV_USER_ID = "11111111-1111-1111-1111-111111111111";

export const DevLogin = () => {
  const { loginDev } = useSession();

  const isDev = import.meta.env.DEV || window.location.hostname === 'localhost';

  if (!isDev) return null;

  return (
    <Button 
      variant="outline" 
      size="sm" 
      onClick={() => loginDev(DEV_USER_ID)}
      className="gap-2 border-dashed border-primary/50 text-primary hover:bg-primary/10"
    >
      <Code className="h-4 w-4" />
      Dev Login
    </Button>
  );
};
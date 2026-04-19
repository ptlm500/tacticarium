import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Spinner } from "@/components/ui/spinner";

export function AuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const token = searchParams.get("token");
    if (token) {
      localStorage.setItem("token", token);
      navigate("/", { replace: true });
    } else {
      navigate("/login", { replace: true });
    }
  }, [searchParams, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
      <div className="flex flex-col items-center gap-3">
        <Spinner size="lg" className="text-primary" />
        <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-primary">
          Establishing uplink
        </p>
      </div>
    </div>
  );
}

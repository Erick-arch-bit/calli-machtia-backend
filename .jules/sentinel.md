## 2023-10-25 - Removed Hardcoded JWT Secret Fallback
**Vulnerability:** A hardcoded `jwtSecret` fallback was present in `src/config.ts` (`"dev-secret-change-in-production"`). If deployed to production without the `JWT_SECRET` environment variable, attackers could forge JWT tokens and gain full unauthorized access.
**Learning:** Fallback secrets meant for local development can accidentally leak into production if environment variables are not strictly validated at startup.
**Prevention:** In configuration logic, explicitly throw an error if critical secrets are not present when the application environment indicates production (`ENV === "production"`).

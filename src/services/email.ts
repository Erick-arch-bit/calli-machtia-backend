/**
 * Servicio de envío de emails mediante Resend.
 * Si RESEND_API_KEY no está configurado, los correos se registran en consola.
 */
import { Resend } from "resend";
import { config } from "../config";

let resend: Resend | null = null;

/** Retorna el cliente de Resend (inicialización lazy) o null si no hay API key */
function getClient(): Resend | null {
  if (resend) return resend;
  if (!config.resendApiKey) {
    console.log("RESEND_API_KEY not configured — emails will be logged to console");
    return null;
  }
  resend = new Resend(config.resendApiKey);
  return resend;
}

/** Envía un email de recuperación de contraseña con un enlace que expira en 1 hora */
export async function sendPasswordResetEmail(email: string, token: string): Promise<void> {
  const resetUrl = `${config.frontendUrl}/reset-password?token=${token}`;

  const r = getClient();

  if (!r) {
    console.log(`[EMAIL] Password reset for ${email}`);
    console.log(`[EMAIL] Token: ${token}`);
    console.log(`[EMAIL] URL: ${resetUrl}`);
    return;
  }

  await r.emails.send({
    from: config.resendFrom,
    to: email,
    subject: "Recuperación de contraseña — Calli Machtia",
    html: `
      <div style="font-family: 'Inter', system-ui, sans-serif; max-width: 480px; margin: 0 auto; padding: 32px 24px;">
        <div style="text-align: center; margin-bottom: 24px;">
          <h1 style="color: #4F46E5; font-size: 24px; margin: 0;">Calli Machtia</h1>
        </div>
        <div style="background: #FFFFFF; border: 1px solid #E2E8F0; border-radius: 16px; padding: 32px;">
          <h2 style="color: #0F172A; font-size: 18px; margin: 0 0 8px;">Recuperación de contraseña</h2>
          <p style="color: #64748B; font-size: 14px; line-height: 1.6; margin: 0 0 24px;">
            Has solicitado restablecer tu contraseña. Haz clic en el botón de abajo para crear una nueva.
          </p>
          <div style="text-align: center; margin-bottom: 24px;">
            <a href="${resetUrl}" style="display: inline-block; background: #4F46E5; color: #FFFFFF; font-size: 14px; font-weight: 600; padding: 12px 32px; border-radius: 12px; text-decoration: none;">
              Restablecer contraseña
            </a>
          </div>
          <p style="color: #94A3B8; font-size: 12px; line-height: 1.5; margin: 0;">
            Este enlace expira en 1 hora. Si no solicitaste este cambio, ignora este correo.
          </p>
        </div>
        <p style="color: #94A3B8; font-size: 12px; text-align: center; margin-top: 24px;">
          — Calli Machtia
        </p>
      </div>
    `,
  });
}

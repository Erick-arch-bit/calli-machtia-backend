import nodemailer from "nodemailer";
import { config } from "../config";

let transporter: nodemailer.Transporter | null = null;

function getTransporter(): nodemailer.Transporter | null {
  if (transporter) return transporter;

  if (!config.smtp.host) {
    console.log("SMTP not configured — emails will be logged to console");
    return null;
  }

  transporter = nodemailer.createTransport({
    host: config.smtp.host,
    port: config.smtp.port,
    secure: config.smtp.port === 465,
    auth: {
      user: config.smtp.user,
      pass: config.smtp.pass,
    },
  });

  return transporter;
}

export async function sendPasswordResetEmail(email: string, token: string): Promise<void> {
  const resetUrl = `${config.frontendUrl}/reset-password?token=${token}`;
  const subject = "Recuperación de contraseña — Calli Machtia";
  const text = `Has solicitado restablecer tu contraseña.\n\nHaz clic en el siguiente enlace para crear una nueva contraseña:\n\n${resetUrl}\n\nEste enlace expira en 1 hora.\n\nSi no solicitaste esto, ignora este correo.\n\n— Calli Machtia`;
  const html = `
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
  `;

  const t = getTransporter();

  if (!t) {
    console.log(`[EMAIL] Password reset for ${email}`);
    console.log(`[EMAIL] Token: ${token}`);
    console.log(`[EMAIL] URL: ${resetUrl}`);
    return;
  }

  await t.sendMail({
    from: config.smtp.from,
    to: email,
    subject,
    text,
    html,
  });
}

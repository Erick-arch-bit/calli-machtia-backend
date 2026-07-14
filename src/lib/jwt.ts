/**
 * Funciones para generación y verificación de tokens JWT.
 * Soporta access tokens (corta duración) y refresh tokens (larga duración).
 */
import jwt from "jsonwebtoken";
import { config } from "../config";

/** Payload del access token JWT */
export interface AccessTokenPayload {
  user_id: string;
  role: string;
}

/** Payload del refresh token JWT */
export interface RefreshTokenPayload {
  user_id: string;
}

/** Convierte una cadena de expiración (ej. "15m", "7d") a segundos */
function parseExpiry(expiry: string): number {
  const match = expiry.match(/^(\d+)([smhd])$/);
  if (!match) return 900;
  const value = parseInt(match[1]);
  const unit = match[2];
  switch (unit) {
    case "s": return value;
    case "m": return value * 60;
    case "h": return value * 3600;
    case "d": return value * 86400;
    default: return 900;
  }
}

/** Genera un access token con el payload y expiración configurados */
export function generateAccessToken(payload: AccessTokenPayload): string {
  return jwt.sign(payload, config.jwtSecret, {
    expiresIn: config.jwtAccessExpiry as any,
  });
}

/** Genera un refresh token con el payload y expiración configurados */
export function generateRefreshToken(payload: RefreshTokenPayload): string {
  return jwt.sign(payload, config.jwtSecret, {
    expiresIn: config.jwtRefreshExpiry as any,
  });
}

/** Verifica un access token y retorna su payload */
export function verifyAccessToken(token: string): AccessTokenPayload {
  return jwt.verify(token, config.jwtSecret) as AccessTokenPayload;
}

/** Verifica un refresh token y retorna su payload */
export function verifyRefreshToken(token: string): RefreshTokenPayload {
  return jwt.verify(token, config.jwtSecret) as RefreshTokenPayload;
}

/** Retorna el TTL del refresh token en segundos según la configuración */
export function getRefreshTokenTTL(): number {
  return parseExpiry(config.jwtRefreshExpiry);
}

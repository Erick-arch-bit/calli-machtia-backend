import jwt from "jsonwebtoken";
import { config } from "../config";

export interface AccessTokenPayload {
  user_id: string;
  role: string;
}

export interface RefreshTokenPayload {
  user_id: string;
}

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

export function generateAccessToken(payload: AccessTokenPayload): string {
  return jwt.sign(payload, config.jwtSecret, {
    expiresIn: config.jwtAccessExpiry as any,
  });
}

export function generateRefreshToken(payload: RefreshTokenPayload): string {
  return jwt.sign(payload, config.jwtSecret, {
    expiresIn: config.jwtRefreshExpiry as any,
  });
}

export function verifyAccessToken(token: string): AccessTokenPayload {
  return jwt.verify(token, config.jwtSecret) as AccessTokenPayload;
}

export function verifyRefreshToken(token: string): RefreshTokenPayload {
  return jwt.verify(token, config.jwtSecret) as RefreshTokenPayload;
}

export function getRefreshTokenTTL(): number {
  return parseExpiry(config.jwtRefreshExpiry);
}

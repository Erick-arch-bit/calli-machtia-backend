/**
 * Cliente de Stripe para procesamiento de pagos.
 * Se inicializa solo si STRIPE_SECRET_KEY está configurada.
 */
import Stripe from "stripe";
import { config } from "../config";

export const stripe = config.stripeSecretKey
  ? new Stripe(config.stripeSecretKey, { apiVersion: "2025-02-24.acacia" })
  : null;

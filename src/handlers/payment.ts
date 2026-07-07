import { Hono } from "hono";
import { eq, and, desc } from "drizzle-orm";
import { z } from "zod";
import { v4 as uuidv4 } from "uuid";
import { db } from "../db/postgres";
import { payments } from "../schema/payments";
import { courses } from "../schema/courses";
import { enrollments } from "../schema/enrollments";
import { users } from "../schema/users";
import { stripe } from "../lib/stripe";
import { config } from "../config";
import { authMiddleware } from "../middleware/auth";
import { badRequest, notFound, internal } from "../lib/errors";
import type { Context } from "hono";

const payment = new Hono();

const createIntentSchema = z.object({
  course_id: z.string().uuid("ID de curso inválido"),
});

const checkoutSchema = z.object({
  course_id: z.string().uuid("ID de curso inválido"),
});

payment.post("/checkout", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const body = await c.req.json();
  const parsed = checkoutSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  if (!stripe) {
    throw internal("Stripe no está configurado");
  }

  const { course_id } = parsed.data;

  const [courseItem] = await db
    .select()
    .from(courses)
    .where(eq(courses.id, course_id))
    .limit(1);

  if (!courseItem) {
    throw notFound("Curso no encontrado");
  }

  const existingEnrollment = await db
    .select()
    .from(enrollments)
    .where(and(eq(enrollments.user_id, userId), eq(enrollments.course_id, course_id)))
    .limit(1);

  if (existingEnrollment.length > 0) {
    throw badRequest("Ya estás inscrito en este curso");
  }

  try {
    const session = await stripe.checkout.sessions.create({
      mode: "payment",
      line_items: [
        {
          price_data: {
            currency: "usd",
            product_data: { name: courseItem.title },
            unit_amount: Math.round(parseFloat(courseItem.price) * 100),
          },
          quantity: 1,
        },
      ],
      metadata: { user_id: userId, course_id },
      success_url: `${config.frontendUrl}/cursos/${courseItem.slug}?pago=exitoso`,
      cancel_url: `${config.frontendUrl}/cursos/${courseItem.slug}?pago=cancelado`,
    });

    return c.json({ data: { url: session.url } });
  } catch (err) {
    console.error("Stripe checkout error:", err);
    throw internal("Error al crear sesión de pago");
  }
});

payment.post("/create-intent", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const body = await c.req.json();
  const parsed = createIntentSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  if (!stripe) {
    throw internal("Stripe no está configurado");
  }

  const { course_id } = parsed.data;

  const courseData = await db
    .select()
    .from(courses)
    .where(eq(courses.id, course_id))
    .limit(1);

  if (courseData.length === 0) {
    throw notFound("Curso no encontrado");
  }

  const courseItem = courseData[0];

  const existingEnrollment = await db
    .select()
    .from(enrollments)
    .where(and(eq(enrollments.user_id, userId), eq(enrollments.course_id, course_id)))
    .limit(1);

  if (existingEnrollment.length > 0) {
    throw badRequest("Ya estás inscrito en este curso");
  }

  try {
    const paymentIntent = await stripe.paymentIntents.create({
      amount: Math.round(parseFloat(courseItem.price) * 100),
      currency: "usd",
      metadata: {
        course_id,
        user_id: userId,
      },
    });

    const paymentId = uuidv4();
    await db.insert(payments).values({
      id: paymentId,
      user_id: userId,
      course_id,
      amount: courseItem.price,
      currency: "usd",
      stripe_payment_id: paymentIntent.id,
      status: "pending",
    });

    return c.json({
      data: {
        payment_intent_id: paymentIntent.id,
        client_secret: paymentIntent.client_secret,
        amount: paymentIntent.amount,
        currency: paymentIntent.currency,
      },
    });
  } catch (err) {
    console.error("Stripe error:", err);
    throw internal("Error al procesar el pago");
  }
});

payment.post("/webhook", async (c: Context) => {
  if (!stripe) {
    throw internal("Stripe no está configurado");
  }

  const signature = c.req.header("stripe-signature");
  if (!signature) {
    throw badRequest("Firma de Stripe requerida");
  }

  let event;
  try {
    const body = await c.req.text();
    event = stripe.webhooks.constructEvent(body, signature, config.stripeWebhookSecret);
  } catch (err) {
    console.error("Stripe webhook signature verification failed:", err);
    return c.json({ error: "Firma inválida" }, 400);
  }

  try {
    if (event.type === "payment_intent.succeeded") {
      const paymentIntent = event.data.object;
      const { course_id, user_id } = paymentIntent.metadata;

      if (course_id && user_id) {
        await db
          .update(payments)
          .set({ status: "completed" })
          .where(eq(payments.stripe_payment_id, paymentIntent.id));

        const existingEnrollment = await db
          .select()
          .from(enrollments)
          .where(and(eq(enrollments.user_id, user_id), eq(enrollments.course_id, course_id)))
          .limit(1);

        if (existingEnrollment.length === 0) {
          await db.insert(enrollments).values({
            id: uuidv4(),
            user_id,
            course_id,
            status: "active",
            progress: "0",
          });
        }
      }
    }

    if (event.type === "payment_intent.payment_failed") {
      const paymentIntent = event.data.object;
      await db
        .update(payments)
        .set({ status: "failed" })
        .where(eq(payments.stripe_payment_id, paymentIntent.id));
    }
  } catch (err) {
    console.error("Webhook processing error:", err);
  }

  return c.json({ received: true });
});

payment.get("/", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;

  const result = await db
    .select({
      id: payments.id,
      amount: payments.amount,
      currency: payments.currency,
      status: payments.status,
      created_at: payments.created_at,
      course: {
        id: courses.id,
        title: courses.title,
        slug: courses.slug,
      },
    })
    .from(payments)
    .leftJoin(courses, eq(payments.course_id, courses.id))
    .where(eq(payments.user_id, userId))
    .orderBy(desc(payments.created_at));

  return c.json({ data: result });
});

export default payment;

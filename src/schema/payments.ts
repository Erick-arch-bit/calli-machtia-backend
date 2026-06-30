import { pgTable, text, timestamp, uuid, decimal } from "drizzle-orm/pg-core";
import { users } from "./users";
import { courses } from "./courses";

export const payments = pgTable("payments", {
  id: uuid("id").defaultRandom().primaryKey(),
  user_id: uuid("user_id")
    .references(() => users.id, { onDelete: "cascade" })
    .notNull(),
  course_id: uuid("course_id")
    .references(() => courses.id, { onDelete: "cascade" })
    .notNull(),
  amount: decimal("amount", { precision: 10, scale: 2 }).notNull(),
  currency: text("currency").default("usd").notNull(),
  stripe_payment_intent_id: text("stripe_payment_intent_id"),
  status: text("status").default("pending").notNull(),
  created_at: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
});

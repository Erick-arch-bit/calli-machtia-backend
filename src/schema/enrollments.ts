import { pgTable, text, timestamp, uuid, decimal, unique } from "drizzle-orm/pg-core";
import { users } from "./users";
import { courses } from "./courses";

export const enrollments = pgTable(
  "enrollments",
  {
    id: uuid("id").defaultRandom().primaryKey(),
    user_id: uuid("user_id")
      .references(() => users.id, { onDelete: "cascade" })
      .notNull(),
    course_id: uuid("course_id")
      .references(() => courses.id, { onDelete: "cascade" })
      .notNull(),
    status: text("status").default("active").notNull(),
    progress: decimal("progress", { precision: 5, scale: 2 }).default("0"),
    enrolled_at: timestamp("enrolled_at", { withTimezone: true }).defaultNow().notNull(),
    completed_at: timestamp("completed_at", { withTimezone: true }),
    created_at: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
    updated_at: timestamp("updated_at", { withTimezone: true }).defaultNow().notNull(),
  },
  (table) => [unique().on(table.user_id, table.course_id)]
);

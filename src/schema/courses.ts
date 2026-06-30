import { pgTable, text, timestamp, uuid, decimal, boolean } from "drizzle-orm/pg-core";
import { users } from "./users";

export const courses = pgTable("courses", {
  id: uuid("id").defaultRandom().primaryKey(),
  instructor_id: uuid("instructor_id")
    .references(() => users.id, { onDelete: "cascade" })
    .notNull(),
  title: text("title").notNull(),
  slug: text("slug").notNull().unique(),
  description: text("description"),
  image_url: text("image_url"),
  price: decimal("price", { precision: 10, scale: 2 }).notNull(),
  category: text("category"),
  tags: text("tags").array(),
  published: boolean("published").default(false),
  seo_title: text("seo_title"),
  seo_description: text("seo_description"),
  deleted_at: timestamp("deleted_at", { withTimezone: true }),
  created_at: timestamp("created_at", { withTimezone: true }).defaultNow().notNull(),
  updated_at: timestamp("updated_at", { withTimezone: true }).defaultNow().notNull(),
});

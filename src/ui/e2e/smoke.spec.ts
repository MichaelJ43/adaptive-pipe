import { test, expect } from "@playwright/test";

test("login page renders", async ({ page }) => {
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: "Sign in" })).toBeVisible();
  await expect(page.getByLabel(/tenant slug/i)).toBeVisible();
});

test("dashboard after login", async ({ page }) => {
  await page.goto("/login");
  await page.getByLabel(/tenant slug/i).fill("demo");
  await page.getByLabel(/^username$/i).fill("admin");
  await page.getByLabel(/^password$/i).fill("admin123");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible({ timeout: 15000 });
});

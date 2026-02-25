import { test, expect } from '@playwright/test'

test.describe.serial('instance flow', () => {
test('create instance', async ({ page }, testInfo) => {
    const instanceName = `test-instance-${testInfo.project.name}`

    await page.goto('/instances/new')

    await page.getByLabel('Name (optional)').fill(instanceName)
    await page.getByLabel('Redis replicas').fill('1')
    await page.getByLabel('Sentinel replicas').fill('1')
    await page.getByRole('button', { name: 'Create Instance' }).click()

    await page.waitForURL('/instances')
    await expect(page.getByText(instanceName, { exact: true })).toBeVisible({ timeout: 15000 })
})

test('update instance', async ({ page }, testInfo) => {
    const instanceName = `test-instance-${testInfo.project.name}`

    await page.goto(`/instances/${instanceName}?namespace=default`)

    await page.getByRole('button', { name: 'Edit instance configuration' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('Redis replicas').fill('2')
    await dialog.getByLabel('Sentinel replicas').fill('2')
    await page.getByRole('button', { name: 'Review changes' }).click()
    await page.getByRole('button', { name: 'Apply' }).click()

    await expect(page.getByText('Instance updated successfully')).toBeVisible()
})

test('delete instance', async ({ page }, testInfo) => {
    const instanceName = `test-instance-${testInfo.project.name}`

    await page.goto(`/instances/${instanceName}?namespace=default`)

    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Yes' }).click();

    await page.waitForURL('/instances')
    await expect(page.getByText('Instance deleted successfully')).toBeVisible();
    await expect(page.getByText(instanceName, { exact: true })).not.toBeVisible();
})
})

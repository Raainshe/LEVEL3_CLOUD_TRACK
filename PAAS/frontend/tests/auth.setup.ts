import {test as setup, expect} from '@playwright/test'

const authFile = 'playwright/.auth/user.json'

setup('authenticate', async ({page}) => {
    await page.goto('/')

    await page.getByLabel('Email').fill('ryanbwgt@gmail.com')
    await page.getByLabel('Password').fill('123')
    await page.getByRole('button', { name: 'Login' }).click()

    await expect(page).toHaveURL('/instances');
    await expect(page.getByRole('heading', { name: 'Redis Instances' })).toBeVisible();

    await page.context().storageState({ path: authFile });
});

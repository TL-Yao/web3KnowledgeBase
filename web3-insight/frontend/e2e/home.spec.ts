import { test, expect } from '@playwright/test'

test.describe('Homepage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('has correct title', async ({ page }) => {
    await expect(page).toHaveTitle(/Web3/)
  })

  test('displays hero section with heading', async ({ page }) => {
    const heading = page.getByRole('heading', { name: 'Web3 Insight' })
    await expect(heading).toBeVisible()
  })

  test('displays description text', async ({ page }) => {
    const description = page.getByText(/私人 Web3 知识管理系统/)
    await expect(description).toBeVisible()
  })

  test('shows search input', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索知识库或输入问题...')
    await expect(searchInput).toBeVisible()
  })

  test('shows search button', async ({ page }) => {
    const searchButton = page.getByRole('button', { name: '搜索' })
    await expect(searchButton).toBeVisible()
  })

  test('can search for articles', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索知识库或输入问题...')
    await searchInput.fill('以太坊')
    await searchInput.press('Enter')

    // Should navigate to knowledge page with search query
    await expect(page).toHaveURL(/\/knowledge\?q=.*以太坊/)
  })

  test('search button submits the search', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索知识库或输入问题...')
    const searchButton = page.getByRole('button', { name: '搜索' })

    await searchInput.fill('比特币')
    await searchButton.click()

    await expect(page).toHaveURL(/\/knowledge\?q=.*比特币/)
  })

  test('displays quick access cards', async ({ page }) => {
    // Check for the three quick access cards
    await expect(page.getByText('知识库')).toBeVisible()
    await expect(page.getByText('即时研究')).toBeVisible()
    await expect(page.getByText('系统设置')).toBeVisible()
  })

  test('quick access card descriptions are visible', async ({ page }) => {
    await expect(page.getByText('浏览和管理你的 Web3 知识文章')).toBeVisible()
    await expect(page.getByText('输入问题，AI 帮你深入研究')).toBeVisible()
    await expect(page.getByText('配置模型和数据源')).toBeVisible()
  })

  test('knowledge card navigates to knowledge page', async ({ page }) => {
    const knowledgeCard = page.getByRole('link', { name: /知识库/ })
    await knowledgeCard.click()

    await expect(page).toHaveURL('/knowledge')
  })

  test('research card navigates to research page', async ({ page }) => {
    const researchCard = page.getByRole('link', { name: /即时研究/ })
    await researchCard.click()

    await expect(page).toHaveURL('/research')
  })

  test('settings card navigates to admin config page', async ({ page }) => {
    const settingsCard = page.getByRole('link', { name: /系统设置/ })
    await settingsCard.click()

    await expect(page).toHaveURL('/admin/config')
  })

  test('displays footer', async ({ page }) => {
    const footer = page.getByText('Web3 Insight - 本地部署的私人知识管理系统')
    await expect(footer).toBeVisible()
  })

  test('empty search does not navigate', async ({ page }) => {
    const searchInput = page.getByPlaceholder('搜索知识库或输入问题...')
    await searchInput.fill('   ')
    await searchInput.press('Enter')

    // Should stay on homepage
    await expect(page).toHaveURL('/')
  })
})

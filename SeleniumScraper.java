import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.Cookie;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;
import org.openqa.selenium.JavascriptExecutor;

import java.io.*;
import java.time.Duration;
import java.util.StringTokenizer;
import java.util.List;

public class SeleniumScraper {
    public static void main(String[] args) {
        // Set ChromeDriver path
        System.setProperty("webdriver.chrome.driver", "./chromedriver");

        // Initialize ChromeDriver
        WebDriver driver = new ChromeDriver();

        // Navigate to Twitter
        driver.get("https://twitter.com");

        // Load cookies from the file
        try (BufferedReader reader = new BufferedReader(new FileReader("twitter_cookies.data"))) {
            String line;
            while ((line = reader.readLine()) != null) {
                StringTokenizer token = new StringTokenizer(line, ";");
                String name = token.nextToken();
                String value = token.nextToken();
                String domain = token.nextToken();
                String path = token.nextToken();
                String expiry = token.nextToken();
                boolean isSecure = Boolean.parseBoolean(token.nextToken());
                Cookie cookie = new Cookie.Builder(name, value)
                        .domain(domain)
                        .path(path)
                        .isSecure(isSecure)
                        .build();
                driver.manage().addCookie(cookie);
            }
        } catch (IOException e) {
            e.printStackTrace();
        }

        // Navigate to the Twitter search page for #whatsapp
        driver.get("https://twitter.com/search?q=%23whatsapp&src=typed_query");

        // Increase timeout to 30 seconds
        WebDriverWait wait = new WebDriverWait(driver, Duration.ofSeconds(30));

        try (BufferedWriter writer = new BufferedWriter(new FileWriter("tweets.txt"))) {
            // Wait for the page to load completely
            JavascriptExecutor js = (JavascriptExecutor) driver;
            wait.until(webDriver -> js.executeScript("return document.readyState").equals("complete"));

            // Scroll down to ensure dynamic content loads
            js.executeScript("window.scrollTo(0, document.body.scrollHeight);");
            Thread.sleep(10000); // Wait for 2 seconds to allow content to load

            // Wait for the presence of articles
            wait.until(ExpectedConditions.presenceOfElementLocated(By.cssSelector("article")));

            // Find and print tweet texts
            List<WebElement> tweets = driver.findElements(By.cssSelector("article"));
            for (WebElement tweet : tweets) {
                String tweetText = tweet.getText();
                System.out.println("Tweet text: " + tweetText);
                writer.write(tweetText);
                writer.newLine();
            }

        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            // Close the browser
            driver.quit();
        }
    }
}

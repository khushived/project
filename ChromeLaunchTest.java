package selenium;
import org.openqa.selenium.WebDriver;
importorg.openqa.selenium.chrome.ChromeDriver;
public class ChromeLaunchTest {
    public static void main(String[] args)
       System.setProperty("webdriver.chrome.driver","C:\\Drivers\\chromedriver.exe");
       WebDriver driver= new ChromeDriver();

       driver.manage().window().maximize();
       driver.get("https://www.google.com");
}

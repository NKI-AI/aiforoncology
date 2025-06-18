#include <gtest/gtest.h>
#include <opencv2/opencv.hpp>
#include <vector>

// Helper function to generate a test image with a circle
cv::Mat CreateTestImage(int height, int width, cv::Point center, int radius) {
  cv::Mat image = cv::Mat::zeros(height, width, CV_8UC3);
  cv::circle(image, center, radius, cv::Scalar(255, 255, 255), 2);
  return image;
}

// Google Test
TEST(OpenCVTest, FindContours) {
  // Create a test image
  int height = 500, width = 500;
  cv::Point center(250, 250);
  int radius = 100;
  cv::Mat image = CreateTestImage(height, width, center, radius);

  // Convert the image to grayscale
  cv::Mat gray_image;
  cv::cvtColor(image, gray_image, cv::COLOR_BGR2GRAY);

  // Find contours
  std::vector<std::vector<cv::Point>> contours;
  cv::findContours(gray_image, contours, cv::RETR_EXTERNAL,
                   cv::CHAIN_APPROX_SIMPLE);

  // Assert that we found exactly one contour (the circle)
  ASSERT_EQ(contours.size(), 1);

  // Draw the contours on the image
  cv::drawContours(image, contours, -1, cv::Scalar(0, 255, 0), cv::FILLED);

  // Check that the contour is approximately circular
  double contour_area = cv::contourArea(contours[0]);
  double expected_area = CV_PI * radius * radius;
  EXPECT_NEAR(contour_area, expected_area,
              expected_area * 0.1);  // Allow 10% error

  // Print debug information
  std::cout << "Number of contours: " << contours.size() << std::endl;
  std::cout << "Contour area: " << contour_area << std::endl;
}

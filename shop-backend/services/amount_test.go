package services

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestAmountCalculation 测试金额计算的准确性
func TestAmountCalculation(t *testing.T) {
	// 测试场景1: 基本乘法计算
	price := decimal.NewFromFloat(19.99)
	quantity := 3
	expected := decimal.NewFromFloat(59.97)
	result := price.Mul(decimal.NewFromInt(int64(quantity)))

	if !result.Equal(expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试场景2: 多次加法计算
	amount1 := decimal.NewFromFloat(0.1)
	amount2 := decimal.NewFromFloat(0.2)
	expectedSum := decimal.NewFromFloat(0.3)
	resultSum := amount1.Add(amount2)

	if !resultSum.Equal(expectedSum) {
		t.Errorf("Expected %s, got %s", expectedSum, resultSum)
	}

	// 测试场景3: 复杂计算
	price1 := decimal.NewFromFloat(10.50)
	quantity1 := 2
	price2 := decimal.NewFromFloat(15.75)
	quantity2 := 3
	expectedTotal := decimal.NewFromFloat(67.25) // 10.50*2 + 15.75*3 = 21 + 47.25 = 68.25? 等等，让我重新计算
	// 正确计算：10.50*2 = 21, 15.75*3 = 47.25, 总和是 68.25
	expectedTotal = decimal.NewFromFloat(68.25)
	resultTotal := price1.Mul(decimal.NewFromInt(int64(quantity1))).Add(price2.Mul(decimal.NewFromInt(int64(quantity2))))

	if !resultTotal.Equal(expectedTotal) {
		t.Errorf("Expected %s, got %s", expectedTotal, resultTotal)
	}

	// 测试场景4: 小数精度测试
	priceSmall := decimal.NewFromFloat(0.01)
	quantityLarge := 100
	expectedSmallTotal := decimal.NewFromFloat(1.00)
	resultSmallTotal := priceSmall.Mul(decimal.NewFromInt(int64(quantityLarge)))

	if !resultSmallTotal.Equal(expectedSmallTotal) {
		t.Errorf("Expected %s, got %s", expectedSmallTotal, resultSmallTotal)
	}
}

// TestAmountComparison 测试金额比较
func TestAmountComparison(t *testing.T) {
	amount1 := decimal.NewFromFloat(10.00)
	amount2 := decimal.NewFromFloat(10.000)
	amount3 := decimal.NewFromFloat(10.01)

	// 测试相等比较
	if !amount1.Equal(amount2) {
		t.Errorf("Expected %s to be equal to %s", amount1, amount2)
	}

	// 测试大于比较
	if !amount3.GreaterThan(amount1) {
		t.Errorf("Expected %s to be greater than %s", amount3, amount1)
	}

	// 测试小于比较
	if !amount1.LessThan(amount3) {
		t.Errorf("Expected %s to be less than %s", amount1, amount3)
	}
}
